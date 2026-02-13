package repositories

import (
	"context"
	"database/sql"
	"errors"

	"autoimport-kz/internal/models"
)

var ErrNotFound = errors.New("not found")

// UserRepository handles user persistence.
type UserRepository interface {
	Create(ctx context.Context, user models.User) (int64, error)
	ByEmail(ctx context.Context, email string) (models.User, error)
	ByID(ctx context.Context, id int64) (models.User, error)
	UpsertAdmin(ctx context.Context, name, email, passwordHash string) error
	Count(ctx context.Context) (int, error)
}

type UserRepo struct {
	db *DB
}

func NewUserRepo(db *DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user models.User) (int64, error) {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	var id int64
	err := r.db.Conn.QueryRowContext(ctx,
		`INSERT INTO users (name, email, password_hash, role) VALUES ($1, $2, $3, $4) RETURNING id`,
		user.Name, user.Email, user.PasswordHash, user.Role).Scan(&id)
	return id, err
}

func (r *UserRepo) ByEmail(ctx context.Context, email string) (models.User, error) {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	var u models.User
	err := r.db.Conn.QueryRowContext(ctx,
		`SELECT id, name, email, password_hash, role FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

func (r *UserRepo) ByID(ctx context.Context, id int64) (models.User, error) {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	var u models.User
	err := r.db.Conn.QueryRowContext(ctx,
		`SELECT id, name, email, password_hash, role FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

func (r *UserRepo) UpsertAdmin(ctx context.Context, name, email, passwordHash string) error {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	_, err := r.db.Conn.ExecContext(ctx, `
		INSERT INTO users (name, email, password_hash, role)
		VALUES ($1, $2, $3, 'admin')
		ON CONFLICT (email) DO UPDATE
		SET name = EXCLUDED.name,
			password_hash = EXCLUDED.password_hash,
			role = 'admin'
	`, name, email, passwordHash)
	return err
}

func (r *UserRepo) Count(ctx context.Context) (int, error) {
	ctx, cancel := r.db.WithTimeout(ctx)
	defer cancel()

	var n int
	err := r.db.Conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}
