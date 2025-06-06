package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	GoogleClient   string
	GoogleSecret   string
	GoogleRedirect string
	JWTSecret      string // Secreto para firmar los tokens JWT
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		Port:           os.Getenv("PORT"),
		GoogleClient:   os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleSecret:   os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirect: os.Getenv("GOOGLE_REDIRECT_URL"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
	}
}
