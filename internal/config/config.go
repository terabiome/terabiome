package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Logger LoggerConfig
}

var cfg *Config

func New() *Config {
	cfg := Config{
		Logger: NewLoggerConfig(),
	}
	return &cfg
}

func Get() *Config {
	if cfg == nil {
		cfg = New()
	}

	return cfg
}

// can log fatal here
func loadEnv[T any](key string, required bool) T {
	value, exist := os.LookupEnv(key)
	if !exist && required {
		log.Fatalf("failed to load %s: not exist\n", key)
	}

	var anchor T
	switch anchorPtr := any(&anchor).(type) {
	case *string:
		*anchorPtr = value
	case *int:
		intVal, err := strconv.Atoi(value)
		if err != nil {
			log.Fatalf("failed to load %s: failed to cast to int: %v\n", key, err)
		}
		*anchorPtr = intVal
	case *bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			log.Fatalf("failed to load %s: failed to cast to bool: %v\n", key, err)
		}
		*anchorPtr = boolVal
	}

	return anchor
}
