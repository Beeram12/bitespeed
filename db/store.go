package db

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type Store struct {
	DB *sqlx.DB
}

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

func (s *Store) Close() error {
	if s == nil || s.DB == nil {
		return nil
	}
	return s.DB.Close()
}
