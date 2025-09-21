package db

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/jmoiron/sqlx"
)

func Migrate(db *sqlx.DB) {
	files, err := filepath.Glob("db/migrations/*.sql")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatal(err)
		}

		_, err = db.Exec(string(data))
		if err != nil {
			log.Fatalf("Migration %s failed: %v", file, err)
		}

		log.Println("Applied migration:", file)
	}
}
