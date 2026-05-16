package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
)

func ConnectToPostgres(connString string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return db, nil
}

func RunMigrations(db *sql.DB) error {

	migrationsSQL, err := ioutil.ReadFile("db/migrations/000001_create_files_table.down.sql")
	if err != nil {
		return fmt.Errorf("не удалось прочитать файл миграций: %v", err)
	}

	_, err = db.Exec(string(migrationsSQL))
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Printf("Successfully run migrations")

	return nil
}
