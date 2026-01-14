package config

import (
	"log"
	"os"
)

type Config struct {
	ServiceName string
	Env         string
	HTTPPort    string
	LogLevel    string
	DataDir     string
}

/*
Why SREs Care:
Config is externalized
Service is 12-factor compliant
Easy to override in containers & Kubernetes
*/
func Load() Config {
	cfg := Config{
		ServiceName: getEnv("SERVICE_NAME", "tiny"),
		Env:         getEnv("SERVICE_ENV", "dev"),
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		DataDir:     getEnv("DATA_DIR", "/data"),
	}

	log.Printf("config loaded: %+v", cfg)
	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
