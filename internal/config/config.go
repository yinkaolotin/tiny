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

	StorageBackend string
	DBHost         string
	DBPort         string
	DBName         string
	DBUser         string
	DBPassword     string
}

/*
Why SREs Care:
Config is externalized
Service is 12-factor compliant
Easy to override in containers & Kubernetes
*/
func Load() Config {
	cfg := Config{
		ServiceName:    getEnv("SERVICE_NAME", "tiny"),
		Env:            getEnv("SERVICE_ENV", "dev"),
		HTTPPort:       getEnv("HTTP_PORT", "8080"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		DataDir:        getEnv("DATA_DIR", "/data"),
		StorageBackend: getEnv("STORAGE_BACKEND", "file"),
		DBHost:         getEnv("DB_HOST", ""),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBName:         getEnv("DB_NAME", ""),
		DBUser:         getEnv("DB_USER", ""),
		DBPassword:     getEnv("DB_PASSWORD", ""),
	}

	log.Printf("config loaded: %+v", cfg)
	return cfg
}

func (c Config) DatabaseURL() string {
	return "postgres://" + c.DBUser + ":" + c.DBPassword + "@" + c.DBHost + ":" + c.DBPort + "/" + c.DBName + "?sslmode=disable"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
