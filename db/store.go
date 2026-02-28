package db

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type Store struct {
	DB *sqlx.DB
}

// NewPostgresStore creates and configures a sqlx-backed Store using the provided DSN.
// It establishes the connection and applies basic pool settings.
func NewPostgresStore(dsn string) (*Store, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	return &Store{
		DB: db,
	}, nil
}

// Close releases the underlying database connection resources, if any.
func (s *Store) Close() error {
	if s == nil || s.DB == nil {
		return nil
	}
	return s.DB.Close()
}
