package repositories

import (
	"context"
	"database/sql"
	"errors"

	"autoimport-kz/internal/models"
)

type CarRepository interface {
	List(ctx context.Context) ([]models.Car, error)
	ByID(ctx context.Context, id int64) (models.Car, error)
	AdminList(ctx context.Context) ([]models.Car, error)
	Create(ctx context.Context, car models.Car) (int64, error)
	Update(ctx context.Context, car models.Car) error
	Delete(ctx context.Context, id int64) error
	Count(ctx context.Context) (int, error)
}

type CarRepo struct {
	db *DB
}

func NewCarRepo(db *DB) *CarRepo {
	return &CarRepo{db: db}
}

func (r *CarRepo) List(ctx context.Context) ([]models.Car, error) {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	rows, err := r.db.Conn.QueryContext(ctx,
		`SELECT c.id, c.brand, c.model, c.year, c.engine_volume, c.engine_type, c.base_price, c.country_id, co.name, COALESCE(c.image_url, '')
		 FROM cars c
		 JOIN countries co ON co.id = c.country_id
		 ORDER BY c.brand, c.model`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Car
	for rows.Next() {
		var c models.Car
		if err := rows.Scan(&c.ID, &c.Brand, &c.Model, &c.Year, &c.EngineVolume, &c.EngineType, &c.BasePrice, &c.CountryID, &c.CountryName, &c.ImageURL); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

func (r *CarRepo) AdminList(ctx context.Context) ([]models.Car, error) {
	return r.List(ctx)
}

func (r *CarRepo) ByID(ctx context.Context, id int64) (models.Car, error) {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	var c models.Car
	err := r.db.Conn.QueryRowContext(ctx,
		`SELECT c.id, c.brand, c.model, c.year, c.engine_volume, c.engine_type, c.base_price, c.country_id, co.name, COALESCE(c.image_url, '')
		 FROM cars c
		 JOIN countries co ON co.id = c.country_id
		 WHERE c.id = $1`, id,
	).Scan(&c.ID, &c.Brand, &c.Model, &c.Year, &c.EngineVolume, &c.EngineType, &c.BasePrice, &c.CountryID, &c.CountryName, &c.ImageURL)
	if errors.Is(err, sql.ErrNoRows) {
		return c, ErrNotFound
	}
	return c, err
}

func (r *CarRepo) Create(ctx context.Context, car models.Car) (int64, error) {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	var id int64
	err := r.db.Conn.QueryRowContext(ctx,
		`INSERT INTO cars (brand, model, year, engine_volume, engine_type, base_price, country_id, image_url)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		car.Brand, car.Model, car.Year, car.EngineVolume, car.EngineType, car.BasePrice, car.CountryID, car.ImageURL,
	).Scan(&id)
	return id, err
}

func (r *CarRepo) Update(ctx context.Context, car models.Car) error {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	res, err := r.db.Conn.ExecContext(ctx,
		`UPDATE cars SET brand=$1, model=$2, year=$3, engine_volume=$4, engine_type=$5, base_price=$6, country_id=$7, image_url=$8 WHERE id=$9`,
		car.Brand, car.Model, car.Year, car.EngineVolume, car.EngineType, car.BasePrice, car.CountryID, car.ImageURL, car.ID,
	)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CarRepo) Delete(ctx context.Context, id int64) error {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	res, err := r.db.Conn.ExecContext(ctx, `DELETE FROM cars WHERE id=$1`, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CarRepo) Count(ctx context.Context) (int, error) {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	var n int
	err := r.db.Conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM cars`).Scan(&n)
	return n, err
}
