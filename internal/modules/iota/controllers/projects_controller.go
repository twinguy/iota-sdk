package controllers

import (
	"fmt"
	"github.com/a-h/templ"
	"github.com/gorilla/mux"
	"github.com/iota-agency/iota-erp/internal/application"
	"github.com/iota-agency/iota-erp/internal/domain/aggregates/project"
	"github.com/iota-agency/iota-erp/internal/modules/shared"
	"github.com/iota-agency/iota-erp/internal/modules/shared/middleware"
	"github.com/iota-agency/iota-erp/internal/presentation/templates/pages/projects"
	"github.com/iota-agency/iota-erp/internal/services"
	"github.com/iota-agency/iota-erp/internal/types"
	"net/http"

	"github.com/iota-agency/iota-erp/internal/presentation/mappers"
	"github.com/iota-agency/iota-erp/internal/presentation/viewmodels"
	"github.com/iota-agency/iota-erp/pkg/composables"
)

type ProjectsController struct {
	app            *application.Application
	projectService *services.ProjectService
	basePath       string
}

func NewProjectsController(app *application.Application) shared.Controller {
	return &ProjectsController{
		app:      app,
		basePath: "/projects",
	}
}

func (c *ProjectsController) Register(r *mux.Router) {
	router := r.PathPrefix(c.basePath).Subrouter()
	router.Use(middleware.RequireAuthorization())
	router.HandleFunc("", c.List).Methods(http.MethodGet)
	router.HandleFunc("", c.Create).Methods(http.MethodPost)
	router.HandleFunc("/{id:[0-9]+}", c.GetEdit).Methods(http.MethodGet)
	router.HandleFunc("/{id:[0-9]+}", c.PostEdit).Methods(http.MethodPost)
	router.HandleFunc("/{id:[0-9]+}", c.Delete).Methods(http.MethodDelete)
	router.HandleFunc("/new", c.GetNew).Methods(http.MethodGet)
}

func (c *ProjectsController) List(w http.ResponseWriter, r *http.Request) {
	pageCtx, err := composables.UsePageCtx(
		r,
		types.NewPageData("Projects.Meta.List.Title", ""),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	params := composables.UsePaginated(r)
	projectEntities, err := c.projectService.GetPaginated(r.Context(), params.Limit, params.Offset, []string{})
	if err != nil {
		http.Error(w, "Error retrieving projects", http.StatusInternalServerError)
		return
	}
	viewProjects := make([]*viewmodels.Project, len(projectEntities))
	for i, entity := range projectEntities {
		viewProjects[i] = mappers.ProjectToViewModel(entity)
	}
	isHxRequest := len(r.Header.Get("Hx-Request")) > 0
	props := &projects.IndexPageProps{
		PageContext: pageCtx,
		Projects:    viewProjects,
		NewURL:      c.basePath + "/new",
	}
	if isHxRequest {
		templ.Handler(projects.ProjectsTable(props), templ.WithStreaming()).ServeHTTP(w, r)
	} else {
		templ.Handler(projects.Index(props), templ.WithStreaming()).ServeHTTP(w, r)
	}
}

func (c *ProjectsController) GetEdit(w http.ResponseWriter, r *http.Request) {
	id, err := shared.ParseID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pageCtx, err := composables.UsePageCtx(
		r,
		types.NewPageData("Projects.Meta.Edit.Title", ""),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	entity, err := c.projectService.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Error retrieving project", http.StatusInternalServerError)
		return
	}
	props := &projects.EditPageProps{
		PageContext: pageCtx,
		Project:     mappers.ProjectToViewModel(entity),
		Errors:      map[string]string{},
		SaveURL:     fmt.Sprintf("%s/%d", c.basePath, id),
		DeleteURL:   fmt.Sprintf("%s/%d", c.basePath, id),
	}
	templ.Handler(projects.Edit(props), templ.WithStreaming()).ServeHTTP(w, r)
}

func (c *ProjectsController) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := shared.ParseID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := c.projectService.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	shared.Redirect(w, r, c.basePath)
}

func (c *ProjectsController) PostEdit(w http.ResponseWriter, r *http.Request) {
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
		if _, err := c.projectService.Delete(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case shared.FormActionSave:
		dto := project.UpdateDTO{} //nolint:exhaustruct
		var pageCtx *types.PageContext
		pageCtx, err = composables.UsePageCtx(
			r,
			types.NewPageData("Projects.Meta.Edit.Title", ""),
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
			if err := c.projectService.Update(r.Context(), id, &dto); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			entity, err := c.projectService.GetByID(r.Context(), id)
			if err != nil {
				http.Error(w, "Error retrieving project", http.StatusInternalServerError)
				return
			}
			props := &projects.EditPageProps{
				PageContext: pageCtx,
				Project:     mappers.ProjectToViewModel(entity),
				Errors:      errorsMap,
				SaveURL:     fmt.Sprintf("%s/%d", c.basePath, id),
				DeleteURL:   fmt.Sprintf("%s/%d", c.basePath, id),
			}
			templ.Handler(projects.EditForm(props), templ.WithStreaming()).ServeHTTP(w, r)
			return
		}
	}
	shared.Redirect(w, r, c.basePath)
}

func (c *ProjectsController) GetNew(w http.ResponseWriter, r *http.Request) {
	pageCtx, err := composables.UsePageCtx(
		r,
		types.NewPageData("Projects.Meta.New.Title", ""),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	props := &projects.CreatePageProps{
		PageContext: pageCtx,
		Errors:      map[string]string{},
		Project:     mappers.ProjectToViewModel(&project.Project{}), //nolint:exhaustruct
		SaveURL:     c.basePath,
	}
	templ.Handler(projects.New(props), templ.WithStreaming()).ServeHTTP(w, r)
}

func (c *ProjectsController) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dto := project.CreateDTO{} //nolint:exhaustruct
	if err := shared.Decoder.Decode(&dto, r.Form); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pageCtx, err := composables.UsePageCtx(
		r, types.NewPageData("Projects.Meta.New.Title", ""),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if errorsMap, ok := dto.Ok(pageCtx.UniTranslator); !ok {
		fmt.Println(errorsMap)
		props := &projects.CreatePageProps{
			PageContext: pageCtx,
			Errors:      errorsMap,
			Project:     mappers.ProjectToViewModel(&project.Project{}), //nolint:exhaustruct
			SaveURL:     c.basePath,
		}
		templ.Handler(projects.CreateForm(props), templ.WithStreaming()).ServeHTTP(w, r)
		return
	}

	if err := c.projectService.Create(r.Context(), &dto); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	shared.Redirect(w, r, c.basePath)
}
