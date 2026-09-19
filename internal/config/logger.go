package config

type LoggerConfig struct {
	Level  string
	Format string
}

func NewLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Level:  loadEnv[string]("LOGGER_LEVEL", false),
		Format: loadEnv[string]("LOGGER_FORMAT", false),
	}
}
