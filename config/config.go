package config

import (
	"fmt"
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
	DatabaseURL    string // URL de conexión a la base de datos
	DBHost         string // Host de la base de datos
	DBPort         string // Puerto de la base de datos
	DBUser         string // Usuario de la base de datos
	DBPassword     string // Contraseña de la base de datos
	DBName         string // Nombre de la base de datos
	DBSSLMode      string // Modo SSL de la base de datos
	FrontendURL    string // URL del frontend para redirección
}

func buildDatabaseURL(cfg *Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBSSLMode,
		cfg.FrontendURL,
	)
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := &Config{
		Port:           os.Getenv("PORT"),
		GoogleClient:   os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleSecret:   os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirect: os.Getenv("GOOGLE_REDIRECT_URL"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		DBHost:         os.Getenv("DB_HOST"),
		DBPort:         os.Getenv("DB_PORT"),
		DBUser:         os.Getenv("DB_USER"),
		DBPassword:     os.Getenv("DB_PASSWORD"),
		DBName:         os.Getenv("DB_NAME"),
		DBSSLMode:      os.Getenv("DB_SSL_MODE"),
		FrontendURL:    os.Getenv("FrontendURL"),
	}

	// Construir la URL de la base de datos
	cfg.DatabaseURL = buildDatabaseURL(cfg)

	return cfg
}
