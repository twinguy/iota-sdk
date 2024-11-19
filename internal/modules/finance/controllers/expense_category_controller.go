package controllers

import (
	"github.com/a-h/templ"
	"github.com/gorilla/mux"
	"github.com/iota-agency/iota-sdk/internal/application"
	expense_categories2 "github.com/iota-agency/iota-sdk/internal/modules/finance/templates/pages/expense_categories"
	"github.com/iota-agency/iota-sdk/internal/modules/shared"
	"github.com/iota-agency/iota-sdk/internal/modules/shared/middleware"
	"github.com/iota-agency/iota-sdk/internal/services"
	"github.com/iota-agency/iota-sdk/internal/types"
	"net/http"

	category "github.com/iota-agency/iota-sdk/internal/domain/aggregates/expense_category"
	"github.com/iota-agency/iota-sdk/internal/presentation/mappers"
	"github.com/iota-agency/iota-sdk/internal/presentation/viewmodels"
	"github.com/iota-agency/iota-sdk/pkg/composables"
)

type ExpenseCategoriesController struct {
	app                    *application.Application
	currencyService        *services.CurrencyService
	expenseCategoryService *services.ExpenseCategoryService
	basePath               string
}

func NewExpenseCategoriesController(app *application.Application) shared.Controller {
	return &ExpenseCategoriesController{
		app:                    app,
		currencyService:        app.Service(services.CurrencyService{}).(*services.CurrencyService),
		expenseCategoryService: app.Service(services.ExpenseCategoryService{}).(*services.ExpenseCategoryService),
		basePath:               "/finance/expense-categories",
	}
}

func (c *ExpenseCategoriesController) Register(r *mux.Router) {
	router := r.PathPrefix(c.basePath).Subrouter()
	router.Use(middleware.RequireAuthorization())
	router.HandleFunc("", c.List).Methods(http.MethodGet)
	router.HandleFunc("", c.Create).Methods(http.MethodPost)
	router.HandleFunc("/{id:[0-9]+}", c.GetEdit).Methods(http.MethodGet)
	router.HandleFunc("/{id:[0-9]+}", c.PostEdit).Methods(http.MethodPost)
	router.HandleFunc("/{id:[0-9]+}", c.Delete).Methods(http.MethodDelete)
	router.HandleFunc("/new", c.GetNew).Methods(http.MethodGet)
}

func (c *ExpenseCategoriesController) viewModelCurrencies(r *http.Request) ([]*viewmodels.Currency, error) {
	currencies, err := c.currencyService.GetAll(r.Context())
	if err != nil {
		return nil, err
	}
	viewCurrencies := make([]*viewmodels.Currency, len(currencies))
	for i, currency := range currencies {
		viewCurrencies[i] = mappers.CurrencyToViewModel(currency)
	}
	return viewCurrencies, nil
}

func (c *ExpenseCategoriesController) List(w http.ResponseWriter, r *http.Request) {
	pageCtx, err := composables.UsePageCtx(
		r,
		types.NewPageData("ExpenseCategories.Meta.List.Title", ""),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	params := composables.UsePaginated(r)
	categories, err := c.expenseCategoryService.GetPaginated(r.Context(), params.Limit, params.Offset, []string{})
	if err != nil {
		http.Error(w, "Error retrieving expense categories", http.StatusInternalServerError)
		return
	}
	viewCategories := make([]*viewmodels.ExpenseCategory, len(categories))
	for i, entity := range categories {
		viewCategories[i] = mappers.ExpenseCategoryToViewModel(entity)
	}
	isHxRequest := len(r.Header.Get("Hx-Request")) > 0
	props := &expense_categories2.IndexPageProps{
		PageContext: pageCtx,
		Categories:  viewCategories,
	}
	if isHxRequest {
		templ.Handler(expense_categories2.CategoriesTable(props), templ.WithStreaming()).ServeHTTP(w, r)
	} else {
		templ.Handler(expense_categories2.Index(props), templ.WithStreaming()).ServeHTTP(w, r)
	}
}

func (c *ExpenseCategoriesController) GetEdit(w http.ResponseWriter, r *http.Request) {
	id, err := shared.ParseID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pageCtx, err := composables.UsePageCtx(
		r,
		types.NewPageData("ExpenseCategories.Meta.Edit.Title", ""),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	entity, err := c.expenseCategoryService.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Error retrieving expense category", http.StatusInternalServerError)
		return
	}
	currencies, err := c.viewModelCurrencies(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	props := &expense_categories2.EditPageProps{
		PageContext: pageCtx,
		Category:    mappers.ExpenseCategoryToViewModel(entity),
		Currencies:  currencies,
		Errors:      map[string]string{},
	}
	templ.Handler(expense_categories2.Edit(props), templ.WithStreaming()).ServeHTTP(w, r)
}

func (c *ExpenseCategoriesController) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := shared.ParseID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := c.expenseCategoryService.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	shared.Redirect(w, r, c.basePath)
}

func (c *ExpenseCategoriesController) PostEdit(w http.ResponseWriter, r *http.Request) {
	id, err := shared.ParseID(r)
	if err != nil {
		http.Error(w, "Error parsing id", http.StatusInternalServerError)
		return
	}
	action := shared.FormAction(r.FormValue("_action"))
	if !action.IsValid() {
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return
	}
	r.Form.Del("_action")

	switch action {
	case shared.FormActionDelete:
		if _, err := c.expenseCategoryService.Delete(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case shared.FormActionSave:
		dto := category.UpdateDTO{} //nolint:exhaustruct
		var pageCtx *types.PageContext
		pageCtx, err = composables.UsePageCtx(
			r,
			types.NewPageData("ExpenseCategories.Meta.Edit.Title", ""),
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := shared.Decoder.Decode(&dto, r.Form); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		errorsMap, ok := dto.Ok(pageCtx.UniTranslator)
		if ok {
			if err := c.expenseCategoryService.Update(r.Context(), id, &dto); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			entity, err := c.expenseCategoryService.GetByID(r.Context(), id)
			if err != nil {
				http.Error(w, "Error retrieving expense category", http.StatusInternalServerError)
				return
			}
			currencies, err := c.viewModelCurrencies(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			props := &expense_categories2.EditPageProps{
				PageContext: pageCtx,
				Category:    mappers.ExpenseCategoryToViewModel(entity),
				Currencies:  currencies,
				Errors:      errorsMap,
			}
			templ.Handler(expense_categories2.EditForm(props), templ.WithStreaming()).ServeHTTP(w, r)
			return
		}
	}
	shared.Redirect(w, r, c.basePath)
}

func (c *ExpenseCategoriesController) GetNew(w http.ResponseWriter, r *http.Request) {
	pageCtx, err := composables.UsePageCtx(
		r,
		types.NewPageData("ExpenseCategories.Meta.New.Title", ""),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	currencies, err := c.viewModelCurrencies(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	props := &expense_categories2.CreatePageProps{
		PageContext: pageCtx,
		Currencies:  currencies,
		Errors:      map[string]string{},
		Category:    category.CreateDTO{}, //nolint:exhaustruct
		PostPath:    c.basePath,
	}
	templ.Handler(expense_categories2.New(props), templ.WithStreaming()).ServeHTTP(w, r)
}

func (c *ExpenseCategoriesController) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dto := category.CreateDTO{} //nolint:exhaustruct
	if err := shared.Decoder.Decode(&dto, r.Form); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pageCtx, err := composables.UsePageCtx(
		r, types.NewPageData("ExpenseCategories.Meta.New.Title", ""),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if errorsMap, ok := dto.Ok(pageCtx.UniTranslator); !ok {
		currencies, err := c.viewModelCurrencies(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		props := &expense_categories2.CreatePageProps{
			PageContext: pageCtx,
			Currencies:  currencies,
			Errors:      errorsMap,
			Category:    dto,
			PostPath:    c.basePath,
		}
		templ.Handler(expense_categories2.CreateForm(props), templ.WithStreaming()).ServeHTTP(w, r)
		return
	}

	if err := c.expenseCategoryService.Create(r.Context(), &dto); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	shared.Redirect(w, r, c.basePath)
}
