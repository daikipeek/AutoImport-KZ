package repositories

import (
	"context"
	"time"

	"autoimport-kz/internal/models"
)

type CalculationRepository interface {
	Create(ctx context.Context, calc models.Calculation) (int64, error)
	AddWarning(ctx context.Context, calculationID int64, message string) error
	HistoryByUser(ctx context.Context, userID int64) ([]models.Calculation, error)
	WarningsByCalculation(ctx context.Context, calculationID int64) ([]models.Warning, error)
	Count(ctx context.Context) (int, error)
	CountByUser(ctx context.Context, userID int64) (int, error)
}

type CalculationRepo struct {
	db *DB
}

func NewCalculationRepo(db *DB) *CalculationRepo {
	return &CalculationRepo{db: db}
}

func (r *CalculationRepo) Create(ctx context.Context, calc models.Calculation) (int64, error) {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	var id int64
	err := r.db.Conn.QueryRowContext(ctx,
		`INSERT INTO calculations (user_id, car_id, total_price, created_at) VALUES ($1, $2, $3, $4) RETURNING id`,
		calc.UserID, calc.CarID, calc.TotalPrice, time.Now().UTC(),
	).Scan(&id)
	return id, err
}

func (r *CalculationRepo) AddWarning(ctx context.Context, calculationID int64, message string) error {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	_, err := r.db.Conn.ExecContext(ctx,
		`INSERT INTO warnings (calculation_id, message) VALUES ($1, $2)`,
		calculationID, message,
	)
	return err
}

func (r *CalculationRepo) HistoryByUser(ctx context.Context, userID int64) ([]models.Calculation, error) {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	rows, err := r.db.Conn.QueryContext(ctx, `
		SELECT cal.id, cal.user_id, cal.car_id, cal.total_price, cal.created_at,
		       c.id, c.brand, c.model, c.year, c.engine_volume, c.engine_type, c.base_price, c.country_id,
		       co.id, co.name, co.customs_rate, co.delivery_cost
		FROM calculations cal
		JOIN cars c ON c.id = cal.car_id
		JOIN countries co ON co.id = c.country_id
		WHERE cal.user_id = $1
		ORDER BY cal.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Calculation
	for rows.Next() {
		var cal models.Calculation
		if err := rows.Scan(
			&cal.ID, &cal.UserID, &cal.CarID, &cal.TotalPrice, &cal.CreatedAt,
			&cal.Car.ID, &cal.Car.Brand, &cal.Car.Model, &cal.Car.Year, &cal.Car.EngineVolume, &cal.Car.EngineType, &cal.Car.BasePrice, &cal.Car.CountryID,
			&cal.Country.ID, &cal.Country.Name, &cal.Country.CustomsRate, &cal.Country.DeliveryCost,
		); err != nil {
			return nil, err
		}
		items = append(items, cal)
	}
	return items, rows.Err()
}

func (r *CalculationRepo) WarningsByCalculation(ctx context.Context, calculationID int64) ([]models.Warning, error) {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	rows, err := r.db.Conn.QueryContext(ctx,
		`SELECT id, calculation_id, message FROM warnings WHERE calculation_id = $1`, calculationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Warning
	for rows.Next() {
		var w models.Warning
		if err := rows.Scan(&w.ID, &w.CalculationID, &w.Message); err != nil {
			return nil, err
		}
		items = append(items, w)
	}
	return items, rows.Err()
}

func (r *CalculationRepo) Count(ctx context.Context) (int, error) {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	var n int
	err := r.db.Conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM calculations`).Scan(&n)
	return n, err
}

func (r *CalculationRepo) CountByUser(ctx context.Context, userID int64) (int, error) {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	var n int
	err := r.db.Conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM calculations WHERE user_id = $1`, userID).Scan(&n)
	return n, err
}
