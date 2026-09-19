package config

type LoggerConfig struct {
	Level     string
	Format    string
	Component string
}

func NewLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Level:     loadEnv[string]("LOGGER_LEVEL", false),
		Format:    loadEnv[string]("LOGGER_FORMAT", false),
		Component: loadEnv[string]("LOGGER_COMPONENT", false),
	}
}
