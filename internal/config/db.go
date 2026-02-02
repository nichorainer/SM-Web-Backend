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
    connStr := os.Getenv("DATABASE_URL")
    if connStr == "" {
        log.Fatal("DATABASE_URL is not set")
    }

    log.Printf("Connecting to DB with string: %s", connStr)

    config, err := pgxpool.ParseConfig(connStr)
    if err != nil {
        log.Fatalf("Unable to parse database config: %v", err)
    }

    pool, err := pgxpool.New(context.Background(), config.ConnString())
    if err != nil {
        log.Fatalf("Unable to connect to database: %v", err)
    }
    db = pool
}

// GetDB returns the database pool
func GetDB() *pgxpool.Pool {
    return db
}