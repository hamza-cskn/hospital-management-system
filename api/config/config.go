package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI          string
	DatabaseName      string
	JWTSecret         string
	TokenExpiration   time.Duration
	ServerPort        string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	return &Config{
		MongoURI:        getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DatabaseName:    getEnv("DB_NAME", "hospital_management"),
		JWTSecret:       getEnv("JWT_SECRET", "your-secret-key"),
		TokenExpiration: 24 * time.Hour,
		ServerPort:      getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
} 