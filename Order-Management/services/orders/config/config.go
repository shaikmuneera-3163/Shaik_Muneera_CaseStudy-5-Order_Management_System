package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port       string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		Port:       getEnv("ORDER_SERVICE_PORT", "8082"),
		DBHost:     getEnv("ORDER_DB_HOST", "localhost"),
		DBPort:     getEnv("ORDER_DB_PORT", "5432"),
		DBUser:     getEnv("ORDER_DB_USER", "postgres"),
		DBPassword: os.Getenv("ORDER_DB_PASSWORD"),
		DBName:     getEnv("ORDER_DB_NAME", "orderdb"),
	}
}

func (c Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
