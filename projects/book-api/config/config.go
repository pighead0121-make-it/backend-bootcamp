package config

import (
	"errors"
	"os"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
}

func Load() (Config, error) {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")

	if user == "" {
		return Config{}, errors.New("DB_USER is required")
	}
	if password == "" {
		return Config{}, errors.New("DB_PASSWORD is required")
	}
	if host == "" {
		return Config{}, errors.New("DB_HOST is required")
	}
	if port == "" {
		return Config{}, errors.New("DB_PORT is required")
	}
	if name == "" {
		return Config{}, errors.New("DB_NAME is required")
	}

	config := Config{
		DBUser:     user,
		DBPassword: password,
		DBHost:     host,
		DBPort:     port,
		DBName:     name,
	}

	return config, nil
}
