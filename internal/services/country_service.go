package services

import (
	"context"

	"autoimport-kz/internal/models"
	"autoimport-kz/internal/repositories"
)

type CountryService struct {
	repo *repositories.CountryRepo
}

func NewCountryService(repo *repositories.CountryRepo) *CountryService {
	return &CountryService{repo: repo}
}

func (s *CountryService) List(ctx context.Context) ([]models.Country, error) {
	return s.repo.List(ctx)
}

func (s *CountryService) ByID(ctx context.Context, id int64) (models.Country, error) {
	return s.repo.ByID(ctx, id)
}

func (s *CountryService) Create(ctx context.Context, c models.Country) (int64, error) {
	if c.Name == "" {
		return 0, ErrInvalidInput
	}
	return s.repo.Create(ctx, c)
}

func (s *CountryService) Update(ctx context.Context, c models.Country) error {
	if c.ID == 0 || c.Name == "" {
		return ErrInvalidInput
	}
	return s.repo.Update(ctx, c)
}

func (s *CountryService) Delete(ctx context.Context, id int64) error {
	if id == 0 {
		return ErrInvalidInput
	}
	return s.repo.Delete(ctx, id)
}
