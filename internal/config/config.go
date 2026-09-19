package config

type Config struct {
	Logger LoggerConfig
}

var cfg *Config

func New() *Config {
	cfg := Config{}
	return &cfg
}

func Get() *Config {
	if cfg == nil {
		cfg = New()
	}

	return cfg
}
