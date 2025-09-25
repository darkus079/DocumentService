package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort    string
	MongoURI      string
	MongoDatabase string
	RedisAddr     string
	MaxFileSize   int64
}

func LoadConfig() *Config {
	return &Config{
		ServerPort:    getEnv("SERVER_PORT", "8000"),
		MongoURI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase: getEnv("MONGO_DATABASE", "documentservice"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		MaxFileSize:   getEnvAsInt64("MAX_FILE_SIZE", 50*1024*1024),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}
