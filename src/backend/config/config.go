package config

import (
	"errors"
	"os"
)

type Config struct {
	DbPassword string
	DbUser     string
	DbHost     string
	DbPort     string
	DbName     string
}

func GetConfig() (*Config, error) {
	cfg := Config{}

	var exists bool
	cfg.DbPassword, exists = os.LookupEnv("POSTGRES_PASSWORD")
	if !exists {
		return nil, errors.New("POSTGRES_PASSWORD is not set")
	}

	cfg.DbUser, exists = os.LookupEnv("POSTGRES_USER")
	if !exists {
		return nil, errors.New("POSTGRES_USER is not set")
	}

	cfg.DbHost, exists = os.LookupEnv("POSTGRES_HOST")
	if !exists {
		return nil, errors.New("POSTGRES_HOST is not set")
	}

	cfg.DbPort, exists = os.LookupEnv("POSTGRES_PORT")
	if !exists {
		return nil, errors.New("POSTGRES_PORT is not set")
	}

	cfg.DbName, exists = os.LookupEnv("DB_NAME")
	if !exists {
		return nil, errors.New("DB_NAME is not set")
	}

	return &cfg, nil

}
