package config

import (
	"context"
    "log"
    "os"

    "github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

// InitDB initializes the database connection pool
func InitDB() {
    dsn := os.Getenv("GOOSE_DBSTRING")
    pool, err := pgxpool.New(context.Background(), dsn)
    if err != nil {
        log.Fatalf("Unable to connect to database: %v", err)
    }
    db = pool
}

// GetDB returns the database pool
func GetDB() *pgxpool.Pool {
    return db
}
