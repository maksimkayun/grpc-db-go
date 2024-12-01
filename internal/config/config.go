package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	Database DatabaseConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

func Load() Config {
	// Загружаем .env файл
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
	}

	return Config{
		Database: DatabaseConfig{
			Host:     getEnvOrDefault("PG_HOST", "localhost"),
			Port:     getEnvOrDefault("PG_PORT", "5432"),
			User:     getEnvOrDefault("PG_USER", "postgres"),
			Password: getEnvOrDefault("PG_PASSWORD", "postgres"),
			Database: getEnvOrDefault("PG_DATABASE_NAME", "note"),
		},
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
