package models

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

type Config struct {
}

func (c *Config) LoadFromLocalPath(path string) error {
	yamlReader, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to read file at path '%s': %w", path, err)
	}

	if err = yaml.NewDecoder(yamlReader).Decode(&c); err != nil {
		return fmt.Errorf("failed to parse decode file: %w", err)
	}

	return nil
}
