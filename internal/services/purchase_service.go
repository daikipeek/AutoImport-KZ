package services

import (
	"context"
	"errors"

	"autoimport-kz/internal/models"
	"autoimport-kz/internal/repositories"
)

type PurchaseService struct {
	cars      *repositories.CarRepo
	purchases *repositories.PurchaseRepo
}

func NewPurchaseService(cars *repositories.CarRepo, purchases *repositories.PurchaseRepo) *PurchaseService {
	return &PurchaseService{cars: cars, purchases: purchases}
}

func (s *PurchaseService) Buy(ctx context.Context, userID, carID int64) (int64, error) {
	if userID <= 0 || carID <= 0 {
		return 0, ErrInvalidInput
	}

	car, err := s.cars.ByID(ctx, carID)
	if err != nil {
		return 0, err
	}

	return s.purchases.Create(ctx, userID, carID, car.BasePrice)
}

func (s *PurchaseService) MyPurchases(ctx context.Context, userID int64) ([]models.Purchase, error) {
	if userID <= 0 {
		return nil, ErrInvalidInput
	}
	return s.purchases.ListByUser(ctx, userID)
}

func (s *PurchaseService) GetByID(ctx context.Context, purchaseID int64) (models.Purchase, error) {
	if purchaseID <= 0 {
		return models.Purchase{}, ErrInvalidInput
	}
	return s.purchases.ByID(ctx, purchaseID)
}

func (s *PurchaseService) MarkPaid(ctx context.Context, userID, purchaseID int64) error {
	p, err := s.purchases.ByID(ctx, purchaseID)
	if err != nil {
		return err
	}
	if p.UserID != userID {
		return errors.New("forbidden")
	}
	if p.Status != "created" {
		return errors.New("purchase cannot be paid in current status")
	}
	return s.purchases.UpdateStatus(ctx, purchaseID, "paid")
}

func (s *PurchaseService) MarkDelivered(ctx context.Context, purchaseID int64) error {
	p, err := s.purchases.ByID(ctx, purchaseID)
	if err != nil {
		return err
	}
	if p.Status != "paid" {
		return errors.New("purchase must be paid before delivery")
	}
	return s.purchases.UpdateStatus(ctx, purchaseID, "delivered")
}
