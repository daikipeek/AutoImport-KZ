package services

import (
	"context"

	"autoimport-kz/internal/models"
	"autoimport-kz/internal/repositories"
)

type CalculationService struct {
	cars      *repositories.CarRepo
	countries *repositories.CountryRepo
	calcs     *repositories.CalculationRepo
}

func NewCalculationService(cars *repositories.CarRepo, countries *repositories.CountryRepo, calcs *repositories.CalculationRepo) *CalculationService {
	return &CalculationService{cars: cars, countries: countries, calcs: calcs}
}

func (s *CalculationService) Calculate(ctx context.Context, userID, carID, countryID int64, currencyRate float64) (models.Calculation, []string, error) {
	car, err := s.cars.ByID(ctx, carID)
	if err != nil {
		return models.Calculation{}, nil, err
	}

	country := models.Country{}
	if countryID != 0 {
		country, err = s.countries.ByID(ctx, countryID)
		if err != nil {
			return models.Calculation{}, nil, err
		}
	} else {
		country, err = s.countries.ByID(ctx, car.CountryID)
		if err != nil {
			return models.Calculation{}, nil, err
		}
	}

	base := car.BasePrice
	if currencyRate > 0 {
		base = base * currencyRate
	}
	customs := base * country.CustomsRate
	serviceFee := base * 0.01
	recycling := 150.0
	if car.EngineVolume > 3000 {
		recycling = 400
	} else if car.EngineVolume > 2000 {
		recycling = 250
	}
	if car.Year < 2015 {
		recycling += 100
	}

	total := base + country.DeliveryCost + customs + serviceFee + recycling

	warnings := []string{}
	if car.EngineVolume > 2500 {
		warnings = append(warnings, "This car is expensive to import due to engine volume")
	}
	if car.Year < 2010 {
		warnings = append(warnings, "Older vehicles may have higher fees and risks")
	}

	calc := models.Calculation{
		UserID:     userID,
		CarID:      car.ID,
		TotalPrice: total,
		Car:        car,
		Country:    country,
		Breakdown: models.Breakdown{
			BasePrice:    base,
			DeliveryCost: country.DeliveryCost,
			CustomsDuty:  customs,
			ServiceFee:   serviceFee,
			RecyclingFee: recycling,
			TotalPrice:   total,
		},
	}

	return calc, warnings, nil
}

func (s *CalculationService) Save(ctx context.Context, calc models.Calculation, warnings []string) (models.Calculation, error) {
	id, err := s.calcs.Create(ctx, calc)
	if err != nil {
		return models.Calculation{}, err
	}
	calc.ID = id

	for _, w := range warnings {
		if err := s.calcs.AddWarning(ctx, id, w); err != nil {
			return models.Calculation{}, err
		}
	}

	return calc, nil
}

func (s *CalculationService) History(ctx context.Context, userID int64) ([]models.Calculation, error) {
	items, err := s.calcs.HistoryByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	for i := range items {
		warnings, err := s.calcs.WarningsByCalculation(ctx, items[i].ID)
		if err != nil {
			return nil, err
		}
		items[i].Warnings = warnings
	}

	return items, nil
}
