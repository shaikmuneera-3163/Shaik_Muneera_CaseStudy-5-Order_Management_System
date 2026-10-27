package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port       string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
	JWTExpiry  int
}

func Load() Config {
	_ = godotenv.Load()

	expiry, err := strconv.Atoi(os.Getenv("JWT_EXPIRY_MINUTES"))
	if err != nil || expiry < 1 {
		expiry = 60
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET is missing from .env")
	}

	return Config{
		Port:       getEnv("AUTH_SERVICE_PORT", "8083"),
		DBHost:     getEnv("AUTH_DB_HOST", "localhost"),
		DBPort:     getEnv("AUTH_DB_PORT", "5432"),
		DBUser:     getEnv("AUTH_DB_USER", "postgres"),
		DBPassword: os.Getenv("AUTH_DB_PASSWORD"),
		DBName:     getEnv("AUTH_DB_NAME", "authdb"),
		JWTSecret:  secret,
		JWTExpiry:  expiry,
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
