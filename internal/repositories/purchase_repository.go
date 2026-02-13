package repositories

import (
	"context"
	"database/sql"
	"errors"

	"autoimport-kz/internal/models"
)

type PurchaseRepo struct {
	db *DB
}

func NewPurchaseRepo(db *DB) *PurchaseRepo {
	return &PurchaseRepo{db: db}
}

func (r *PurchaseRepo) Create(ctx context.Context, userID, carID int64, price float64) (int64, error) {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	var id int64
	err := r.db.Conn.QueryRowContext(ctx, `
		INSERT INTO purchases (user_id, car_id, price, status, created_at)
		VALUES ($1, $2, $3, 'created', NOW())
		RETURNING id
	`, userID, carID, price).Scan(&id)
	return id, err
}

func (r *PurchaseRepo) ListByUser(ctx context.Context, userID int64) ([]models.Purchase, error) {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	rows, err := r.db.Conn.QueryContext(ctx, `
		SELECT p.id, p.user_id, p.car_id, p.price, p.status, p.created_at,
		       u.name,
		       c.id, c.brand, c.model, c.year, c.engine_volume, c.engine_type, c.base_price, c.country_id, co.name, COALESCE(c.image_url, '')
		FROM purchases p
		JOIN users u ON u.id = p.user_id
		JOIN cars c ON c.id = p.car_id
		JOIN countries co ON co.id = c.country_id
		WHERE p.user_id = $1
		ORDER BY p.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Purchase
	for rows.Next() {
		var p models.Purchase
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.CarID, &p.Price, &p.Status, &p.CreatedAt,
			&p.UserName,
			&p.Car.ID, &p.Car.Brand, &p.Car.Model, &p.Car.Year, &p.Car.EngineVolume, &p.Car.EngineType, &p.Car.BasePrice, &p.Car.CountryID, &p.Car.CountryName, &p.Car.ImageURL,
		); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

func (r *PurchaseRepo) ByID(ctx context.Context, id int64) (models.Purchase, error) {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	var p models.Purchase
	err := r.db.Conn.QueryRowContext(ctx, `
		SELECT p.id, p.user_id, p.car_id, p.price, p.status, p.created_at,
		       u.name,
		       c.id, c.brand, c.model, c.year, c.engine_volume, c.engine_type, c.base_price, c.country_id, co.name, COALESCE(c.image_url, '')
		FROM purchases p
		JOIN users u ON u.id = p.user_id
		JOIN cars c ON c.id = p.car_id
		JOIN countries co ON co.id = c.country_id
		WHERE p.id = $1
	`, id).Scan(
		&p.ID, &p.UserID, &p.CarID, &p.Price, &p.Status, &p.CreatedAt,
		&p.UserName,
		&p.Car.ID, &p.Car.Brand, &p.Car.Model, &p.Car.Year, &p.Car.EngineVolume, &p.Car.EngineType, &p.Car.BasePrice, &p.Car.CountryID, &p.Car.CountryName, &p.Car.ImageURL,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return p, ErrNotFound
	}
	return p, err
}

func (r *PurchaseRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	res, err := r.db.Conn.ExecContext(ctx, `UPDATE purchases SET status = $1 WHERE id = $2`, status, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PurchaseRepo) Recent(ctx context.Context, limit int) ([]models.Purchase, error) {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	rows, err := r.db.Conn.QueryContext(ctx, `
		SELECT p.id, p.user_id, p.car_id, p.price, p.status, p.created_at,
		       u.name,
		       c.id, c.brand, c.model, c.year, c.engine_volume, c.engine_type, c.base_price, c.country_id, co.name, COALESCE(c.image_url, '')
		FROM purchases p
		JOIN users u ON u.id = p.user_id
		JOIN cars c ON c.id = p.car_id
		JOIN countries co ON co.id = c.country_id
		ORDER BY p.created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Purchase
	for rows.Next() {
		var p models.Purchase
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.CarID, &p.Price, &p.Status, &p.CreatedAt,
			&p.UserName,
			&p.Car.ID, &p.Car.Brand, &p.Car.Model, &p.Car.Year, &p.Car.EngineVolume, &p.Car.EngineType, &p.Car.BasePrice, &p.Car.CountryID, &p.Car.CountryName, &p.Car.ImageURL,
		); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

func (r *PurchaseRepo) Count(ctx context.Context) (int, error) {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()
	var n int
	err := r.db.Conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM purchases`).Scan(&n)
	return n, err
}

func (r *PurchaseRepo) CountByUser(ctx context.Context, userID int64) (int, error) {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()
	var n int
	err := r.db.Conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM purchases WHERE user_id = $1`, userID).Scan(&n)
	return n, err
}

func (r *PurchaseRepo) AveragePrice(ctx context.Context) (float64, error) {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()
	var avg sql.NullFloat64
	err := r.db.Conn.QueryRowContext(ctx, `SELECT AVG(price) FROM purchases`).Scan(&avg)
	if err != nil {
		return 0, err
	}
	if !avg.Valid {
		return 0, nil
	}
	return avg.Float64, nil
}
