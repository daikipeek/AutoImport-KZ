package handlers

import (
	"net/http"

	"autoimport-kz/internal/middleware"
	"autoimport-kz/internal/models"
	"autoimport-kz/internal/services"
	"github.com/gorilla/sessions"
)

type DashboardHandler struct {
	dashboard   *services.DashboardService
	renderer    *TemplateRenderer
	store       *sessions.CookieStore
	sessionName string
}

func NewDashboardHandler(dashboard *services.DashboardService, renderer *TemplateRenderer, store *sessions.CookieStore, sessionName string) *DashboardHandler {
	return &DashboardHandler{
		dashboard:   dashboard,
		renderer:    renderer,
		store:       store,
		sessionName: sessionName,
	}
}

type dashboardPageData struct {
	Title           string
	ContentTemplate string
	User            models.User
	Summary         models.DashboardSummary
	Flash           string
}

func (h *DashboardHandler) Show(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	var (
		summary models.DashboardSummary
		err     error
	)

	if user.Role == "admin" {
		summary, err = h.dashboard.ForAdmin(r.Context())
	} else {
		summary, err = h.dashboard.ForUser(r.Context(), user.ID)
	}
	if err != nil {
		http.Error(w, "failed to load dashboard", http.StatusInternalServerError)
		return
	}

	h.renderer.Render(w, "dashboard.html", dashboardPageData{
		Title:           "Dashboard",
		ContentTemplate: "dashboard_content",
		User:            user,
		Summary:         summary,
		Flash:           popFlash(w, r, h.store, h.sessionName),
	})
}
