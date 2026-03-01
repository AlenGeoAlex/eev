package main

import (
	config "backend-go/config"
	"errors"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "modernc.org/sqlite"
)

func main() {
	var direction string
	flag.StringVar(&direction, "direction", "up", "migration direction: up or down")
	flag.Parse()

	dbConfig := config.NewDBConfig()
	var connectionURL string = dbConfig.MigrationConnectionString()
	var migrationPath string = dbConfig.DbType.MigrationPath()

	if err := ensureDir(dbConfig.SqlitePath); err != nil {
		log.Fatal(err)
	}

	log.Printf("Running migration in %s direction with connection %s.", direction, connectionURL)
	log.Printf("Migration path: %s", migrationPath)

	m, err := migrate.New(migrationPath, connectionURL)

	if err != nil {
		log.Fatal(err)
		return
	}

	if direction == "up" {
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal(err)
		}
	} else {
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal(err)
		}
	}

	log.Println("Migration finished")
}

func ensureDir(path string) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, 0755)
}
