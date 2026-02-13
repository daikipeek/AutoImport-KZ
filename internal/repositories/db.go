package repositories

import (
	"context"
	"database/sql"
	"time"
)

type DB struct {
	Conn *sql.DB
}

func NewDB(conn *sql.DB) *DB {
	return &DB{Conn: conn}
}

func (d *DB) WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 5*time.Second)
}
