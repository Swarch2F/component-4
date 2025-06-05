package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	GoogleClient string
	GoogleSecret string
	GoogleRedirect string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return &Config{
		Port:         os.Getenv("PORT"),
		GoogleClient: os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirect: os.Getenv("GOOGLE_REDIRECT_URL"),
	}
}