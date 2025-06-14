package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
)

func main() {
	// Obtener variables de entorno
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "authuser")
	dbPass := getEnv("DB_PASSWORD", "authpass")
	dbName := getEnv("DB_NAME", "authdb")

	// Conectar a la base de datos
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)
	
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error conectando a la base de datos: %v", err)
	}
	defer db.Close()

	// Verificar conexión
	if err := db.Ping(); err != nil {
		log.Fatalf("Error haciendo ping a la base de datos: %v", err)
	}

	// Leer y ejecutar archivos de migración
	migrationsDir := "./migrations"
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("Error leyendo directorio de migraciones: %v", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".sql" {
			content, err := ioutil.ReadFile(filepath.Join(migrationsDir, file.Name()))
			if err != nil {
				log.Fatalf("Error leyendo archivo %s: %v", file.Name(), err)
			}

			fmt.Printf("Ejecutando migración: %s\n", file.Name())
			_, err = db.Exec(string(content))
			if err != nil {
				log.Fatalf("Error ejecutando migración %s: %v", file.Name(), err)
			}
			fmt.Printf("Migración %s completada\n", file.Name())
		}
	}

	fmt.Println("Todas las migraciones completadas exitosamente")
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
} 