package config

import (
	"fmt"
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
	// SMS_CONFIG is the one file for a live acceptance process.
	// When it is set, the process cannot start as mock.
	locked := strings.TrimSpace(os.Getenv("SMS_CONFIG"))
	if locked != "" {
		if err := godotenv.Overload(locked); err != nil {
			return nil, fmt.Errorf("SMS_CONFIG: %w", err)
		}
	} else {
		// Project .env must win over inherited shell vars (e.g. medx-core SERVER_PORT).
		_ = godotenv.Overload(".env")
	}
	cfg := &Config{
		ServerPort:          getEnv("SERVER_PORT", "8063"),
		GinMode:             getEnv("GIN_MODE", "debug"),
		SMSProvider:         strings.ToLower(strings.TrimSpace(getEnv("SMS_PROVIDER", "mock"))),
		SynermaxxURL:        strings.TrimSpace(getEnv("SYNERMAXX_SMS_URL", "")),
		SynermaxxAppKey:     strings.TrimSpace(getEnv("SYNERMAXX_SMS_APPKEY", "")),
		SynermaxxOriginator: strings.TrimSpace(getEnv("SYNERMAXX_SMS_ORIGINATOR", "22566100")),
	}
	if locked != "" && cfg.SMSProvider != "synermaxx" {
		return nil, fmt.Errorf("SMS_CONFIG requires SMS_PROVIDER=synermaxx")
	}
	if locked != "" && cfg.SynermaxxAppKey == "" {
		return nil, fmt.Errorf("SMS_CONFIG requires SYNERMAXX_SMS_APPKEY")
	}
	return cfg, nil
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
