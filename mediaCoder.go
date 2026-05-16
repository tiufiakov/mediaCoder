package mediaCoder

import (
	"log"

	"github.com/tiufiakov/mediaCoder/db"
)

func Start() {
	connstring := "user=postgres password=1234 dbname=mediaCoder host=localhost port=5432 sslmode=disable"

	database, err := db.ConnectToPostgres(connstring)
	if err != nil {
		log.Fatal(err)
	}

	defer database.Close()

	err = db.RunMigrations(database)

}
