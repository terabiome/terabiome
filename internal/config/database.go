package config

type DatabaseConfig struct {
	DSN  string
	Type string // postgresql
}

func NewDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		DSN:  loadEnv[string]("DB_CONNECTION_STRING", true),
		Type: loadEnv[string]("DB_CONNECTION_TYPE", true),
	}
}
