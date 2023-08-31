package dbGen

import (
	"encoding/json"
	"fmt"
	"os"
)

const configPath = "./db-gen.json"

func GetConfig() (*Config, error) {
	config, err := readJsonConfigFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("getting configuration from file: %w", err)
	}

	return config, nil
}

func readJsonConfigFile(path string) (*Config, error) {
	file, err := os.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}

	config := new(Config)

	err = json.Unmarshal(file, config)

	if err != nil {
		return nil, fmt.Errorf("parsing json: %w", err)
	}

	return config, nil
}
