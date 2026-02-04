package config

import (
	"os"
)

type Config struct {
	ServerAddress string
	DatabasePath  string
}

func Load() *Config {
	cfg := &Config{
		ServerAddress: ":8080",
		DatabasePath:  "krizzy.db",
	}

	if addr := os.Getenv("SERVER_ADDRESS"); addr != "" {
		cfg.ServerAddress = addr
	}
	if dbPath := os.Getenv("DATABASE_PATH"); dbPath != "" {
		cfg.DatabasePath = dbPath
	}

	return cfg
}
