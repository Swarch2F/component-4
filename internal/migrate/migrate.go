package migrate

import (
	"database/sql"
	"io/ioutil"
	"log"
	"path/filepath"
)

// RunMigrations ejecuta todos los archivos .sql en el directorio de migraciones sobre la base de datos dada.
func RunMigrations(db *sql.DB, migrationsDir string) error {
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".sql" {
			content, err := ioutil.ReadFile(filepath.Join(migrationsDir, file.Name()))
			if err != nil {
				return err
			}
			log.Printf("Ejecutando migración: %s", file.Name())
			_, err = db.Exec(string(content))
			if err != nil {
				return err
			}
			log.Printf("Migración %s completada", file.Name())
		}
	}
	log.Println("Todas las migraciones completadas exitosamente")
	return nil
}
