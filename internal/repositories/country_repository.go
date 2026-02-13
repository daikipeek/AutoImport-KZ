package repositories

import (
	"context"
	"database/sql"
	"errors"

	"autoimport-kz/internal/models"
)

type CountryRepository interface {
	List(ctx context.Context) ([]models.Country, error)
	Create(ctx context.Context, c models.Country) (int64, error)
	Update(ctx context.Context, c models.Country) error
	Delete(ctx context.Context, id int64) error
	ByID(ctx context.Context, id int64) (models.Country, error)
	Count(ctx context.Context) (int, error)
}

type CountryRepo struct {
	db *DB
}

func NewCountryRepo(db *DB) *CountryRepo {
	return &CountryRepo{db: db}
}

func (r *CountryRepo) List(ctx context.Context) ([]models.Country, error) {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	rows, err := r.db.Conn.QueryContext(ctx, `SELECT id, name, customs_rate, delivery_cost FROM countries ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Country
	for rows.Next() {
		var c models.Country
		if err := rows.Scan(&c.ID, &c.Name, &c.CustomsRate, &c.DeliveryCost); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

func (r *CountryRepo) ByID(ctx context.Context, id int64) (models.Country, error) {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	var c models.Country
	err := r.db.Conn.QueryRowContext(ctx,
		`SELECT id, name, customs_rate, delivery_cost FROM countries WHERE id = $1`, id,
	).Scan(&c.ID, &c.Name, &c.CustomsRate, &c.DeliveryCost)
	if errors.Is(err, sql.ErrNoRows) {
		return c, ErrNotFound
	}
	return c, err
}

func (r *CountryRepo) Create(ctx context.Context, c models.Country) (int64, error) {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	var id int64
	err := r.db.Conn.QueryRowContext(ctx,
		`INSERT INTO countries (name, customs_rate, delivery_cost) VALUES ($1, $2, $3) RETURNING id`,
		c.Name, c.CustomsRate, c.DeliveryCost,
	).Scan(&id)
	return id, err
}

func (r *CountryRepo) Update(ctx context.Context, c models.Country) error {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	res, err := r.db.Conn.ExecContext(ctx,
		`UPDATE countries SET name=$1, customs_rate=$2, delivery_cost=$3 WHERE id=$4`,
		c.Name, c.CustomsRate, c.DeliveryCost, c.ID,
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

func (r *CountryRepo) Delete(ctx context.Context, id int64) error {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	res, err := r.db.Conn.ExecContext(ctx, `DELETE FROM countries WHERE id=$1`, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CountryRepo) Count(ctx context.Context) (int, error) {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	var n int
	err := r.db.Conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM countries`).Scan(&n)
	return n, err
}
