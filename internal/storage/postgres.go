package storage

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	DB *pgxpool.Pool
}

func NewPostgresStore(dsn string) *PostgresStore {
	dbpool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	return &PostgresStore{
		DB: dbpool,
	}
}

func (s *PostgresStore) Close() {
	s.DB.Close()
}
