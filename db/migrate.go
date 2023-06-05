package db

import (
	"log"
	"sync"

	migrate "github.com/rubenv/sql-migrate"
)

var migrateOnce sync.Once

func Migrate() {
	migrateOnce.Do(func() {
		log.Println("Auto migrating...")
		migrations := migrate.FileMigrationSource{
			Dir: "db/migrations",
		}

		_, err := migrate.Exec(DB(), "sqlite3", migrations, migrate.Up)
		if err != nil {
			throwErr(err)
		}
	})
}
