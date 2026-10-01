package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	RedisURL     string
	JWTSecret    string
	GroupAddress string
}

func Load() Config {
	_ = godotenv.Load()
	_ = godotenv.Load("services/location/.env")

	return Config{
		Port:         env("PORT", "50053"),
		RedisURL:     env("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:    env("JWT_SECRET", "rally-dev-jwt-secret"),
		GroupAddress: env("GROUP_ADDRESS", "localhost:50052"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
