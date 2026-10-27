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
		Port:       getEnv("PRODUCT_SERVICE_PORT", "8081"),
		DBHost:     getEnv("PRODUCT_DB_HOST", "localhost"),
		DBPort:     getEnv("PRODUCT_DB_PORT", "5432"),
		DBUser:     getEnv("PRODUCT_DB_USER", "postgres"),
		DBPassword: os.Getenv("PRODUCT_DB_PASSWORD"),
		DBName:     getEnv("PRODUCT_DB_NAME", "productdb"),
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
