package env

import (
  "os"
)

type Config struct {
  DBHost     string
  DBPort     string
  DBUser     string
  DBPassword string
  DBName     string
  JWTSecret  string
}

func Load() Config {
  return Config{
    DBHost:     getEnv("DB_HOST", "localhost"),
    DBPort:     getEnv("DB_PORT", "5432"),
    DBUser:     getEnv("DB_USER", "postgres"),
    DBPassword: getEnv("DB_PASSWORD", "admin"),
    DBName:     getEnv("DB_NAME", "InventoryDB"),
    JWTSecret:  getEnv("JWT_SECRET", "change_this_secret"),
  }
}

func getEnv(key, fallback string) string {
  if v := os.Getenv(key); v != "" {
    return v
  }
  return fallback
}