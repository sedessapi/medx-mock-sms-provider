package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort          string
	GinMode             string
	SMSProvider         string
	SynermaxxURL        string
	SynermaxxAppKey     string
	SynermaxxOriginator string
}

func Load() (*Config, error) {
	// Project .env must win over inherited shell vars (e.g. medx-core SERVER_PORT).
	_ = godotenv.Overload(".env")
	return &Config{
		ServerPort:          getEnv("SERVER_PORT", "8063"),
		GinMode:             getEnv("GIN_MODE", "debug"),
		SMSProvider:         strings.ToLower(strings.TrimSpace(getEnv("SMS_PROVIDER", "mock"))),
		SynermaxxURL:        strings.TrimSpace(getEnv("SYNERMAXX_SMS_URL", "")),
		SynermaxxAppKey:     strings.TrimSpace(getEnv("SYNERMAXX_SMS_APPKEY", "")),
		SynermaxxOriginator: strings.TrimSpace(getEnv("SYNERMAXX_SMS_ORIGINATOR", "22566100")),
	}, nil
}

func (c *Config) LiveSynermaxx() bool {
	return c != nil && c.SMSProvider == "synermaxx"
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
