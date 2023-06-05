package db

import (
	"database/sql"
	"log"

	migrate "github.com/rubenv/sql-migrate"
)

func Migrate() {
	log.Println("Auto migrating...")
	migrations := migrate.FileMigrationSource{
		Dir: "db/migrations",
	}

	db, err := sql.Open("sqlite3", DBName)
	if err != nil {
		throwErr(err)
	}

	_, err = migrate.Exec(db, "sqlite3", migrations, migrate.Up)
	if err != nil {
		throwErr(err)
	}
}
