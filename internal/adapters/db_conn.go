package postgresql

import (
  "database/sql"
  "fmt"
  _ "github.com/lib/pq"
  "time"
  "github.com/yourorg/Back-End/internal/env"
)

func NewDB(cfg env.Config) (*sql.DB, error) {
  dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
    cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
  db, err := sql.Open("postgres", dsn)
  if err != nil {
    return nil, err
  }
  db.SetMaxOpenConns(25)
  db.SetMaxIdleConns(25)
  db.SetConnMaxLifetime(5 * time.Minute)
  if err := db.Ping(); err != nil {
    return nil, err
  }
  return db, nil
}