package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"autoimport-kz/internal/models"
	"autoimport-kz/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
)

type CarHandler struct {
	cars        *services.CarService
	countries   *services.CountryService
	renderer    *TemplateRenderer
	store       *sessions.CookieStore
	sessionName string
}

func NewCarHandler(cars *services.CarService, countries *services.CountryService, renderer *TemplateRenderer, store *sessions.CookieStore, sessionName string) *CarHandler {
	return &CarHandler{
		cars:        cars,
		countries:   countries,
		renderer:    renderer,
		store:       store,
		sessionName: sessionName,
	}
}

type carsListData struct {
	Title           string
	ContentTemplate string
	Cars            []models.Car
}

type carDetailData struct {
	Title           string
	ContentTemplate string
	Car             models.Car
}

type adminCarsData struct {
	Title           string
	ContentTemplate string
	Cars            []models.Car
	Countries       []models.Country
	EditCar         *models.Car
	Error           string
	Flash           string
}

func (h *CarHandler) List(w http.ResponseWriter, r *http.Request) {
	cars, err := h.cars.List(r.Context())
	if err != nil {
		http.Error(w, "failed to load cars", http.StatusInternalServerError)
		return
	}
	h.renderer.Render(w, "cars_list.html", carsListData{Title: "Cars", ContentTemplate: "cars_list_content", Cars: cars})
}

func (h *CarHandler) Detail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		http.NotFound(w, r)
		return
	}
	car, err := h.cars.ByID(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	h.renderer.Render(w, "car_detail.html", carDetailData{Title: "Car Details", ContentTemplate: "car_detail_content", Car: car})
}

func (h *CarHandler) AdminList(w http.ResponseWriter, r *http.Request) {
	cars, err := h.cars.AdminList(r.Context())
	if err != nil {
		http.Error(w, "failed to load cars", http.StatusInternalServerError)
		return
	}
	countries, _ := h.countries.List(r.Context())

	var editCar *models.Car
	if editID := r.URL.Query().Get("edit_id"); editID != "" {
		id, parseErr := strconv.ParseInt(editID, 10, 64)
		if parseErr == nil && id > 0 {
			if car, getErr := h.cars.ByID(r.Context(), id); getErr == nil {
				editCar = &car
			}
		}
	}

	h.renderer.Render(w, "admin_cars.html", adminCarsData{
		Title:           "Admin Cars",
		ContentTemplate: "admin_cars_content",
		Cars:            cars,
		Countries:       countries,
		EditCar:         editCar,
		Flash:           popFlash(w, r, h.store, h.sessionName),
	})
}

func (h *CarHandler) AdminCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.renderAdminCarsError(w, r, "Invalid form")
		return
	}

	car, err := parseCarForm(r)
	if err != nil {
		h.renderAdminCarsError(w, r, "Invalid input")
		return
	}

	imageURL, imageErr := saveUploadedCarImage(r, "image")
	if imageErr != nil {
		h.renderAdminCarsError(w, r, imageErr.Error())
		return
	}
	car.ImageURL = imageURL

	if _, err := h.cars.Create(r.Context(), car); err != nil {
		h.renderAdminCarsError(w, r, err.Error())
		return
	}

	setFlash(w, r, h.store, h.sessionName, "Car created successfully.")
	http.Redirect(w, r, "/admin/cars", http.StatusFound)
}

func (h *CarHandler) AdminUpdateByPost(w http.ResponseWriter, r *http.Request) {
	h.adminUpdate(w, r)
}

func (h *CarHandler) AdminUpdate(w http.ResponseWriter, r *http.Request) {
	h.adminUpdate(w, r)
}

func (h *CarHandler) adminUpdate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	car, err := parseCarForm(r)
	if err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}
	car.ID = id

	existing, err := h.cars.ByID(r.Context(), id)
	if err != nil {
		http.Error(w, "car not found", http.StatusNotFound)
		return
	}

	imageURL, imageErr := saveUploadedCarImage(r, "image")
	if imageErr != nil {
		http.Error(w, imageErr.Error(), http.StatusBadRequest)
		return
	}
	if imageURL == "" {
		imageURL = existing.ImageURL
	}
	car.ImageURL = imageURL

	if err := h.cars.Update(r.Context(), car); err != nil {
		http.Error(w, "update failed", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		setFlash(w, r, h.store, h.sessionName, "Car updated successfully.")
		http.Redirect(w, r, "/admin/cars", http.StatusFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CarHandler) AdminDeleteByPost(w http.ResponseWriter, r *http.Request) {
	h.adminDelete(w, r)
}

func (h *CarHandler) AdminDelete(w http.ResponseWriter, r *http.Request) {
	h.adminDelete(w, r)
}

func (h *CarHandler) adminDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.cars.Delete(r.Context(), id); err != nil {
		http.Error(w, "delete failed", http.StatusInternalServerError)
		return
	}
	if r.Method == http.MethodPost {
		setFlash(w, r, h.store, h.sessionName, "Car deleted successfully.")
		http.Redirect(w, r, "/admin/cars", http.StatusFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CarHandler) renderAdminCarsError(w http.ResponseWriter, r *http.Request, msg string) {
	cars, _ := h.cars.AdminList(r.Context())
	countries, _ := h.countries.List(r.Context())
	h.renderer.Render(w, "admin_cars.html", adminCarsData{
		Title:           "Admin Cars",
		ContentTemplate: "admin_cars_content",
		Cars:            cars,
		Countries:       countries,
		Error:           msg,
		Flash:           popFlash(w, r, h.store, h.sessionName),
	})
}

func parseCarForm(r *http.Request) (models.Car, error) {
	year, err := strconv.Atoi(r.FormValue("year"))
	if err != nil || year <= 0 {
		return models.Car{}, errors.New("invalid year")
	}
	volume, err := strconv.Atoi(r.FormValue("engine_volume"))
	if err != nil || volume <= 0 {
		return models.Car{}, errors.New("invalid engine volume")
	}
	base, err := strconv.ParseFloat(r.FormValue("base_price"), 64)
	if err != nil || base <= 0 {
		return models.Car{}, errors.New("invalid base price")
	}
	countryID, err := strconv.ParseInt(r.FormValue("country_id"), 10, 64)
	if err != nil || countryID <= 0 {
		return models.Car{}, errors.New("invalid country id")
	}

	car := models.Car{
		Brand:        strings.TrimSpace(r.FormValue("brand")),
		Model:        strings.TrimSpace(r.FormValue("model")),
		Year:         year,
		EngineVolume: volume,
		EngineType:   strings.TrimSpace(r.FormValue("engine_type")),
		BasePrice:    base,
		CountryID:    countryID,
	}
	return car, nil
}

func saveUploadedCarImage(r *http.Request, field string) (string, error) {
	file, header, err := r.FormFile(field)
	if err != nil {
		// no uploaded file is valid
		return "", nil
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
	default:
		return "", errors.New("image must be .jpg, .jpeg, .png, or .webp")
	}

	if err := os.MkdirAll(filepath.FromSlash("static/styles/images/cars"), 0755); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("upload-%d%s", time.Now().UnixNano(), ext)
	targetPath := filepath.FromSlash("static/styles/images/cars/" + filename)
	dst, err := os.Create(targetPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := dst.ReadFrom(file); err != nil {
		return "", err
	}

	return "/static/styles/images/cars/" + filename, nil
}
