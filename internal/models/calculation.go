package models

import "time"

type Calculation struct {
	ID         int64
	UserID     int64
	CarID      int64
	TotalPrice float64
	CreatedAt  time.Time

	Car     Car
	Country Country
	Warnings []Warning

	Breakdown Breakdown
}

type Breakdown struct {
	BasePrice     float64
	DeliveryCost  float64
	CustomsDuty   float64
	ServiceFee    float64
	RecyclingFee  float64
	TotalPrice    float64
}
