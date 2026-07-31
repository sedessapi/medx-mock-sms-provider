package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string
	GinMode    string
}

func Load() (*Config, error) {
	// Project .env must win over inherited shell vars (e.g. medx-core SERVER_PORT).
	_ = godotenv.Overload(".env")
	return &Config{
		ServerPort: getEnv("SERVER_PORT", "8063"),
		GinMode:    getEnv("GIN_MODE", "debug"),
	}, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
