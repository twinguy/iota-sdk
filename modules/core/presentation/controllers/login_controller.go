package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/iota-uz/iota-sdk/modules/core/infrastructure/persistence"
	"github.com/iota-uz/iota-sdk/modules/core/services"
	"github.com/iota-uz/iota-sdk/pkg/configuration"
	"github.com/iota-uz/iota-sdk/pkg/middleware"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/iota-uz/iota-sdk/modules/core/presentation/templates/pages/login"
	"github.com/iota-uz/iota-sdk/pkg/application"
	"github.com/iota-uz/iota-sdk/pkg/constants"
	"github.com/iota-uz/iota-sdk/pkg/shared"
	"github.com/iota-uz/iota-sdk/pkg/types"

	"github.com/gorilla/mux"
	"github.com/iota-uz/iota-sdk/pkg/composables"
)

type LoginDTO struct {
	Email    string `validate:"required"`
	Password string `validate:"required"`
}

func (e *LoginDTO) Ok(l ut.Translator) (map[string]string, bool) {
	errorMessages := map[string]string{}
	errs := constants.Validate.Struct(e)
	if errs == nil {
		return errorMessages, true
	}

	for _, err := range errs.(validator.ValidationErrors) {
		errorMessages[err.Field()] = err.Translate(l)
	}
	return errorMessages, len(errorMessages) == 0
}

func NewLoginController(app application.Application) application.Controller {
	return &LoginController{
		app:         app,
		authService: app.Service(services.AuthService{}).(*services.AuthService),
	}
}

type LoginController struct {
	app         application.Application
	authService *services.AuthService
}

func (c *LoginController) Key() string {
	return "/login"
}

func (c *LoginController) Register(r *mux.Router) {
	getRouter := r.PathPrefix("/").Subrouter()
	getRouter.Use(middleware.WithLocalizer(c.app.Bundle()))
	getRouter.HandleFunc("/login", c.Get).Methods(http.MethodGet)
	getRouter.HandleFunc("/oauth/google/callback", c.GoogleCallback)

	setRouter := r.PathPrefix("/login").Subrouter()
	setRouter.Use(
		middleware.WithLocalizer(c.app.Bundle()),
		middleware.WithTransaction(),
	)
	setRouter.HandleFunc("", c.Post).Methods(http.MethodPost)
}

func (c *LoginController) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	pageCtx, err := composables.UsePageCtx(r, types.NewPageData("Login.Meta.Title", ""))
	queryParams := url.Values{
		"next": []string{r.URL.Query().Get("next")},
	}
	if err != nil {
		queryParams.Set("error", pageCtx.T("Errors.Internal"))
		http.Redirect(w, r, fmt.Sprintf("/login?%s", queryParams.Encode()), http.StatusFound)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		queryParams.Set("error", pageCtx.T("Login.Errors.OauthCodeNotFound"))
		http.Redirect(w, r, fmt.Sprintf("/login?%s", queryParams.Encode()), http.StatusFound)
		return
	}
	state := r.URL.Query().Get("state")
	if state == "" {
		queryParams.Set("error", pageCtx.T("Login.Errors.OauthStateNotFound"))
		http.Redirect(w, r, fmt.Sprintf("/login?%s", queryParams.Encode()), http.StatusFound)
		return
	}
	conf := configuration.Use()
	oauthCookie, err := r.Cookie(conf.OauthStateCookieKey)
	if err != nil {
		queryParams.Set("error", pageCtx.T("Login.Errors.OauthStateNotFound"))
		http.Redirect(w, r, fmt.Sprintf("/login?%s", queryParams.Encode()), http.StatusFound)
		return
	}
	if oauthCookie.Value != state {
		queryParams.Set("error", pageCtx.T("Login.Errors.OauthStateInvalid"))
		http.Redirect(w, r, fmt.Sprintf("/login?%s", queryParams.Encode()), http.StatusFound)
		return
	}
	cookie, err := c.authService.CookieGoogleAuthenticate(r.Context(), code)
	if err != nil {
		if errors.Is(err, persistence.ErrUserNotFound) {
			queryParams.Set("error", pageCtx.T("Login.Errors.UserNotFound"))
		} else {
			queryParams.Set("error", pageCtx.T("Errors.Internal"))
		}
		http.Redirect(w, r, fmt.Sprintf("/login?%s", queryParams.Encode()), http.StatusFound)
		return
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (c *LoginController) Get(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	errorsMap, err := composables.UseFlashMap[string, string](w, r, "errorsMap")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	errorMessage, err := composables.UseFlash(w, r, "error")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pageCtx, err := composables.UsePageCtx(r, types.NewPageData("Login.Meta.Title", ""))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	codeURL, err := c.authService.GoogleAuthenticate(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := login.Index(&login.LoginProps{
		PageContext:        pageCtx,
		ErrorsMap:          errorsMap,
		Email:              email,
		ErrorMessage:       string(errorMessage),
		GoogleOAuthCodeURL: codeURL,
	}).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *LoginController) Post(w http.ResponseWriter, r *http.Request) {
	dto, err := composables.UseForm(&LoginDTO{}, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pageCtx, err := composables.UsePageCtx(r, types.NewPageData("Login.Meta.Title", ""))
	if err != nil {
		shared.SetFlash(w, "error", []byte(pageCtx.T("Errors.Internal")))
		http.Redirect(w, r, fmt.Sprintf("/login?next=%s", r.URL.Query().Get("next")), http.StatusFound)
		return
	}
	if errorsMap, ok := dto.Ok(pageCtx.UniTranslator); !ok {
		shared.SetFlashMap(w, "errorsMap", errorsMap)
		http.Redirect(w, r, fmt.Sprintf("/login?email=%s&next=%s", dto.Email, r.URL.Query().Get("next")), http.StatusFound)
		return
	}

	cookie, err := c.authService.CookieAuthenticate(r.Context(), dto.Email, dto.Password)
	if err != nil {
		if errors.Is(err, composables.ErrInvalidPassword) {
			shared.SetFlash(w, "error", []byte(pageCtx.T("Login.Errors.PasswordInvalid")))
		} else {
			shared.SetFlash(w, "error", []byte(pageCtx.T("Errors.Internal")))
		}
		http.Redirect(w, r, fmt.Sprintf("/login?email=%s&next=%s", dto.Email, r.URL.Query().Get("next")), http.StatusFound)
		return
	}

	redirectURL := r.URL.Query().Get("next")
	if redirectURL == "" {
		redirectURL = "/"
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}
