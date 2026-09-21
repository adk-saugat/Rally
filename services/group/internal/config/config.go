package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
}

func Load() Config {
	_ = godotenv.Load()
	_ = godotenv.Load("services/group/.env")

	return Config{
		Port:        env("PORT", "50052"),
		DatabaseURL: env("DATABASE_URL", "postgres://rally:rally@localhost:5434/group?sslmode=disable"),
		JWTSecret:   env("JWT_SECRET", "rally-dev-jwt-secret"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
