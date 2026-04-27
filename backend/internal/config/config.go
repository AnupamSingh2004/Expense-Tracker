package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DBHost         string
	DBPort         int
	DBName         string
	DBUser         string
	DBPassword     string
	DBSSLMode      string
	ServerPort     int
	LogLevel       string
	MigrationsPath string
	AllowedOrigins string
	JWTSecret      string
}

func Load() (*Config, error) {
	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}
	srvPort, err := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_PORT: %w", err)
	}
	return &Config{
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         dbPort,
		DBName:         getEnv("DB_NAME", "expenses"),
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPassword:     getEnv("DB_PASSWORD", ""),
		DBSSLMode:      getEnv("DB_SSLMODE", "disable"),
		ServerPort:     srvPort,
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		MigrationsPath: getEnv("MIGRATIONS_PATH", "./internal/migrations"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "*"),
		JWTSecret:      getEnv("JWT_SECRET", "change-me-in-production"),
	}, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBName, c.DBUser, c.DBPassword, c.DBSSLMode,
	)
}

// PostgresDSN returns the DSN in URL form required by golang-migrate.
func (c *Config) PostgresDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
