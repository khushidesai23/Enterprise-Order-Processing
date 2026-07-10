package config

type Config struct {
	App AppConfig
	DB  DatabaseConfig
	Log LogConfig
}

type AppConfig struct {
	Name string
	Env  string
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type LogConfig struct {
	Level string
}

var cfg *Config

// Get returns the loaded application configuration.
func Get() *Config {
	return cfg
}