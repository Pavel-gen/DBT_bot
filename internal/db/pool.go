package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Connect(ctx context.Context, dsn string) error {
	var err error
	Pool, err = pgxpool.New(ctx, dsn)

	if err != nil {
		return fmt.Errorf("failed to create db pool: %w", err)
	}

	if err := Pool.Ping(ctx); err != nil {
		return fmt.Errorf("Failes to ping db: %w", err);
	}

	log.Println("Database connected")
	return nil
}