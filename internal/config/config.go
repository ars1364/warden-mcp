package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port       string
	DBPath     string
	MasterKey  string // base64-encoded 32-byte AES key
	JWTSecret  string
	Env        string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:      getEnv("WARDEN_PORT", "7070"),
		DBPath:    getEnv("WARDEN_DB_PATH", "./data/warden.db"),
		MasterKey: os.Getenv("WARDEN_MASTER_KEY"),
		JWTSecret: os.Getenv("WARDEN_JWT_SECRET"),
		Env:       getEnv("WARDEN_ENV", "production"),
	}

	if cfg.MasterKey == "" {
		return nil, fmt.Errorf("WARDEN_MASTER_KEY is required (base64-encoded 32-byte key)")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("WARDEN_JWT_SECRET is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
