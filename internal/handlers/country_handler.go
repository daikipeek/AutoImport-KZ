package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"autoimport-kz/internal/models"
	"autoimport-kz/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
)

type CountryHandler struct {
	countries   *services.CountryService
	renderer    *TemplateRenderer
	store       *sessions.CookieStore
	sessionName string
}

func NewCountryHandler(countries *services.CountryService, renderer *TemplateRenderer, store *sessions.CookieStore, sessionName string) *CountryHandler {
	return &CountryHandler{
		countries:   countries,
		renderer:    renderer,
		store:       store,
		sessionName: sessionName,
	}
}

type adminCountriesData struct {
	Title           string
	ContentTemplate string
	Countries       []models.Country
	EditCountry     *models.Country
	Error           string
	Flash           string
}

func (h *CountryHandler) AdminList(w http.ResponseWriter, r *http.Request) {
	countries, err := h.countries.List(r.Context())
	if err != nil {
		http.Error(w, "failed to load countries", http.StatusInternalServerError)
		return
	}

	var editCountry *models.Country
	if editID := r.URL.Query().Get("edit_id"); editID != "" {
		id, parseErr := strconv.ParseInt(editID, 10, 64)
		if parseErr == nil && id > 0 {
			if c, getErr := h.countries.ByID(r.Context(), id); getErr == nil {
				editCountry = &c
			}
		}
	}

	h.renderer.Render(w, "admin_countries.html", adminCountriesData{
		Title:           "Admin Countries",
		ContentTemplate: "admin_countries_content",
		Countries:       countries,
		EditCountry:     editCountry,
		Flash:           popFlash(w, r, h.store, h.sessionName),
	})
}

func (h *CountryHandler) AdminCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.renderCountriesError(w, r, "Invalid form")
		return
	}

	country, err := parseCountryForm(r)
	if err != nil {
		h.renderCountriesError(w, r, "Invalid input")
		return
	}

	_, err = h.countries.Create(r.Context(), country)
	if err != nil {
		h.renderCountriesError(w, r, err.Error())
		return
	}

	setFlash(w, r, h.store, h.sessionName, "Country created successfully.")
	http.Redirect(w, r, "/admin/countries", http.StatusFound)
}

func (h *CountryHandler) AdminUpdateByPost(w http.ResponseWriter, r *http.Request) {
	h.adminUpdate(w, r)
}

func (h *CountryHandler) AdminUpdate(w http.ResponseWriter, r *http.Request) {
	h.adminUpdate(w, r)
}

func (h *CountryHandler) adminUpdate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	country, err := parseCountryForm(r)
	if err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}
	country.ID = id

	if err := h.countries.Update(r.Context(), country); err != nil {
		http.Error(w, "update failed", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		setFlash(w, r, h.store, h.sessionName, "Country updated successfully.")
		http.Redirect(w, r, "/admin/countries", http.StatusFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CountryHandler) AdminDeleteByPost(w http.ResponseWriter, r *http.Request) {
	h.adminDelete(w, r)
}

func (h *CountryHandler) AdminDelete(w http.ResponseWriter, r *http.Request) {
	h.adminDelete(w, r)
}

func (h *CountryHandler) adminDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.countries.Delete(r.Context(), id); err != nil {
		http.Error(w, "delete failed", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		setFlash(w, r, h.store, h.sessionName, "Country deleted successfully.")
		http.Redirect(w, r, "/admin/countries", http.StatusFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CountryHandler) renderCountriesError(w http.ResponseWriter, r *http.Request, msg string) {
	countries, _ := h.countries.List(r.Context())
	h.renderer.Render(w, "admin_countries.html", adminCountriesData{
		Title:           "Admin Countries",
		ContentTemplate: "admin_countries_content",
		Countries:       countries,
		Error:           msg,
		Flash:           popFlash(w, r, h.store, h.sessionName),
	})
}

func parseCountryForm(r *http.Request) (models.Country, error) {
	customs, err := strconv.ParseFloat(r.FormValue("customs_rate"), 64)
	if err != nil || customs < 0 {
		return models.Country{}, errors.New("invalid customs rate")
	}
	delivery, err := strconv.ParseFloat(r.FormValue("delivery_cost"), 64)
	if err != nil || delivery < 0 {
		return models.Country{}, errors.New("invalid delivery cost")
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		return models.Country{}, errors.New("name is required")
	}

	country := models.Country{
		Name:         name,
		CustomsRate:  customs,
		DeliveryCost: delivery,
	}
	return country, nil
}
