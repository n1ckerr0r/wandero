package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	ServerAddr      string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	DBSSLMode       string
	CORSOrigins     []string
	APIWriteKey     string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		ServerAddr:  getEnv("SERVER_ADDR", ":8080"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "chronograph"),
		DBPassword:  getEnv("DB_PASSWORD", "chronograph"),
		DBName:      getEnv("DB_NAME", "chronograph"),
		DBSSLMode:   getEnv("DB_SSLMODE", "disable"),
		APIWriteKey: os.Getenv("API_WRITE_KEY"),
	}

	var err error
	if cfg.ReadTimeout, err = getDuration("HTTP_READ_TIMEOUT", 10*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.WriteTimeout, err = getDuration("HTTP_WRITE_TIMEOUT", 30*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.IdleTimeout, err = getDuration("HTTP_IDLE_TIMEOUT", 60*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.ShutdownTimeout, err = getDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second); err != nil {
		return Config{}, err
	}
	for _, origin := range strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			cfg.CORSOrigins = append(cfg.CORSOrigins, origin)
		}
	}

	if cfg.ServerAddr == "" || cfg.DBHost == "" || cfg.DBPort == "" || cfg.DBUser == "" || cfg.DBName == "" {
		return Config{}, fmt.Errorf("database config is incomplete")
	}
	if cfg.DBSSLMode != "disable" && cfg.DBSSLMode != "allow" && cfg.DBSSLMode != "prefer" && cfg.DBSSLMode != "require" && cfg.DBSSLMode != "verify-ca" && cfg.DBSSLMode != "verify-full" {
		return Config{}, fmt.Errorf("DB_SSLMODE has unsupported value %q", cfg.DBSSLMode)
	}

	return cfg, nil
}

func getDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration", key)
	}
	return parsed, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
