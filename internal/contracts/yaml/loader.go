package yamlcontracts

import (
	"fmt"
	"io"
	"os"

	"go.yaml.in/yaml/v4"
)

type Loader struct {
}

func NewLoader(options ...any) *Loader {
	loader := Loader{}

	// can add extra parameter here
	return &loader
}

func (l Loader) LoadConfigFromLocalPath(path string) (*Config, error) {
	yamlReader, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file at path '%s': %w", path, err)
	}
	defer yamlReader.Close()

	cfgStruct := Config{}
	if err = yaml.NewDecoder(yamlReader).Decode(&cfgStruct); err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to parse decode file: %w", err)
	}

	return &cfgStruct, nil
}
