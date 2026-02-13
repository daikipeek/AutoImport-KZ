package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"autoimport-kz/internal/middleware"
	"autoimport-kz/internal/models"
	"autoimport-kz/internal/services"
)

type CalculatorHandler struct {
	cars      *services.CarService
	countries *services.CountryService
	calcs     *services.CalculationService
	renderer  *TemplateRenderer
}

func NewCalculatorHandler(cars *services.CarService, countries *services.CountryService, calcs *services.CalculationService, renderer *TemplateRenderer) *CalculatorHandler {
	return &CalculatorHandler{cars: cars, countries: countries, calcs: calcs, renderer: renderer}
}

type calculatorData struct {
	Title           string
	ContentTemplate string
	Cars            []models.Car
	Countries       []models.Country
	Result          *models.Calculation
	Warnings        []string
	Error           string
}

type historyData struct {
	Title           string
	ContentTemplate string
	Calculations    []models.Calculation
}

func (h *CalculatorHandler) ShowCalculator(w http.ResponseWriter, r *http.Request) {
	cars, _ := h.cars.List(r.Context())
	countries, _ := h.countries.List(r.Context())
	h.renderer.Render(w, "calculator.html", calculatorData{Title: "Calculator", ContentTemplate: "calculator_content", Cars: cars, Countries: countries})
}

func (h *CalculatorHandler) Calculate(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	cars, _ := h.cars.List(r.Context())
	countries, _ := h.countries.List(r.Context())

	if err := r.ParseForm(); err != nil {
		h.renderer.Render(w, "calculator.html", calculatorData{
			Title:           "Calculator",
			ContentTemplate: "calculator_content",
			Cars:            cars,
			Countries:       countries,
			Error:           "Invalid form",
		})
		return
	}

	carID, err := strconv.ParseInt(r.FormValue("car_id"), 10, 64)
	if err != nil || carID <= 0 {
		h.renderer.Render(w, "calculator.html", calculatorData{
			Title:           "Calculator",
			ContentTemplate: "calculator_content",
			Cars:            cars,
			Countries:       countries,
			Error:           "Invalid car ID",
		})
		return
	}

	countryID := int64(0)
	if rawCountryID := r.FormValue("country_id"); rawCountryID != "" {
		parsedCountryID, parseErr := strconv.ParseInt(rawCountryID, 10, 64)
		if parseErr != nil || parsedCountryID <= 0 {
			h.renderer.Render(w, "calculator.html", calculatorData{
				Title:           "Calculator",
				ContentTemplate: "calculator_content",
				Cars:            cars,
				Countries:       countries,
				Error:           "Invalid country ID",
			})
			return
		}
		countryID = parsedCountryID
	}

	rate, err := parseOptionalRate(r.FormValue("currency_rate"))
	if err != nil {
		h.renderer.Render(w, "calculator.html", calculatorData{
			Title:           "Calculator",
			ContentTemplate: "calculator_content",
			Cars:            cars,
			Countries:       countries,
			Error:           err.Error(),
		})
		return
	}

	calc, warnings, err := h.calcs.Calculate(r.Context(), user.ID, carID, countryID, rate)
	if err != nil {
		h.renderer.Render(w, "calculator.html", calculatorData{
			Title:           "Calculator",
			ContentTemplate: "calculator_content",
			Cars:            cars,
			Countries:       countries,
			Error:           err.Error(),
		})
		return
	}

	calc, err = h.calcs.Save(r.Context(), calc, warnings)
	if err != nil {
		h.renderer.Render(w, "calculator.html", calculatorData{
			Title:           "Calculator",
			ContentTemplate: "calculator_content",
			Cars:            cars,
			Countries:       countries,
			Error:           "Failed to save calculation",
		})
		return
	}

	h.renderer.Render(w, "calculator.html", calculatorData{Title: "Calculator", ContentTemplate: "calculator_content", Cars: cars, Countries: countries, Result: &calc, Warnings: warnings})
}

func (h *CalculatorHandler) History(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	items, err := h.calcs.History(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "failed to load history", http.StatusInternalServerError)
		return
	}
	h.renderer.Render(w, "history.html", historyData{Title: "History", ContentTemplate: "history_content", Calculations: items})
}

func parseOptionalRate(raw string) (float64, error) {
	if raw == "" {
		return 0, nil
	}
	rate, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, errors.New("invalid currency rate")
	}
	if rate <= 0 {
		return 0, errors.New("currency rate must be greater than zero")
	}
	return rate, nil
}
