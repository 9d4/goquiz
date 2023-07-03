package entity

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/asdine/storm"
)

var (
	DBName = "goquiz.db"

	db       *storm.DB
	openOnce sync.Once

	AllowAutoSeed = true
)

func DB() *storm.DB {
	return db
}

func Open(purge ...bool) (err error) {
	if len(purge) > 0 && purge[0] {
		f, err := os.Open(DBName)
		if err == nil {
			f.Close()

			var confirmation string

			fmt.Print("Old database found. Do you want to remove it? If no then auto seed will be canceled, old data will be used instead (y/n): ")
			fmt.Scanln(&confirmation)

			if strings.HasPrefix(strings.ToLower(confirmation), "y") {
				log.Print("removing...")
				os.Remove(DBName)
			} else {
				AllowAutoSeed = false
			}
		}

	}

	openOnce.Do(func() { err = open() })
	return
}

func open() error {
	d, err := storm.Open(DBName, storm.Batch())
	if err != nil {
		return err
	}

	db = d
	initialize()
	return nil
}

func initialize() {
	DB().Init(&User{})
	DB().Init(&Question{})
	DB().Init(&Choice{})
}
