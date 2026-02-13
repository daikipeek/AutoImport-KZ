package models

import "time"

type Purchase struct {
	ID        int64
	UserID    int64
	UserName  string
	CarID     int64
	Price     float64
	Status    string
	CreatedAt time.Time

	Car Car
}
