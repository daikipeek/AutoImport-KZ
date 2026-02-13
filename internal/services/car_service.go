package services

import (
	"context"

	"autoimport-kz/internal/models"
	"autoimport-kz/internal/repositories"
)

type CarService struct {
	repo *repositories.CarRepo
}

func NewCarService(repo *repositories.CarRepo) *CarService {
	return &CarService{repo: repo}
}

func (s *CarService) List(ctx context.Context) ([]models.Car, error) {
	return s.repo.List(ctx)
}

func (s *CarService) ByID(ctx context.Context, id int64) (models.Car, error) {
	return s.repo.ByID(ctx, id)
}

func (s *CarService) AdminList(ctx context.Context) ([]models.Car, error) {
	return s.repo.AdminList(ctx)
}

func (s *CarService) Create(ctx context.Context, c models.Car) (int64, error) {
	if c.Brand == "" || c.Model == "" || c.Year == 0 || c.EngineVolume == 0 || c.BasePrice == 0 || c.CountryID == 0 {
		return 0, ErrInvalidInput
	}
	return s.repo.Create(ctx, c)
}

func (s *CarService) Update(ctx context.Context, c models.Car) error {
	if c.ID == 0 {
		return ErrInvalidInput
	}
	return s.repo.Update(ctx, c)
}

func (s *CarService) Delete(ctx context.Context, id int64) error {
	if id == 0 {
		return ErrInvalidInput
	}
	return s.repo.Delete(ctx, id)
}
