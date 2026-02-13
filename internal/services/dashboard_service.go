package services

import (
	"context"

	"autoimport-kz/internal/models"
	"autoimport-kz/internal/repositories"
)

type DashboardService struct {
	users     *repositories.UserRepo
	cars      *repositories.CarRepo
	countries *repositories.CountryRepo
	calcs     *repositories.CalculationRepo
	purchases *repositories.PurchaseRepo
}

func NewDashboardService(
	users *repositories.UserRepo,
	cars *repositories.CarRepo,
	countries *repositories.CountryRepo,
	calcs *repositories.CalculationRepo,
	purchases *repositories.PurchaseRepo,
) *DashboardService {
	return &DashboardService{
		users:     users,
		cars:      cars,
		countries: countries,
		calcs:     calcs,
		purchases: purchases,
	}
}

func (s *DashboardService) ForAdmin(ctx context.Context) (models.DashboardSummary, error) {
	var out models.DashboardSummary
	var err error

	if out.UsersCount, err = s.users.Count(ctx); err != nil {
		return out, err
	}
	if out.CarsCount, err = s.cars.Count(ctx); err != nil {
		return out, err
	}
	if out.CountriesCount, err = s.countries.Count(ctx); err != nil {
		return out, err
	}
	if out.CalculationsCount, err = s.calcs.Count(ctx); err != nil {
		return out, err
	}
	if out.PurchasesCount, err = s.purchases.Count(ctx); err != nil {
		return out, err
	}
	if out.AveragePurchaseSum, err = s.purchases.AveragePrice(ctx); err != nil {
		return out, err
	}
	if out.RecentPurchases, err = s.purchases.Recent(ctx, 8); err != nil {
		return out, err
	}
	return out, nil
}

func (s *DashboardService) ForUser(ctx context.Context, userID int64) (models.DashboardSummary, error) {
	var out models.DashboardSummary
	var err error

	if out.CalculationsCount, err = s.calcs.CountByUser(ctx, userID); err != nil {
		return out, err
	}
	if out.PurchasesCount, err = s.purchases.CountByUser(ctx, userID); err != nil {
		return out, err
	}
	return out, nil
}
