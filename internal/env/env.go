package env

import (
	"os"
	// "github.com/joho/godotenv"
)

// Config holds environment configuration used across the app.
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
}

// Load reads environment variables and returns a Config with sensible defaults.
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

// func Load() error {
//     return godotenv.Load()
// }

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// GetString returns the environment variable value for key or fallback if empty.
func GetString(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
