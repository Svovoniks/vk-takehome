package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	ApiUrl string `json:"api_url"`
}

func GetConfig() (*Config, error) {
	data, err := os.ReadFile("cfg.json")
	if err != nil {
		return nil, err
	}

	var cfg Config

	if err = json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil

}
