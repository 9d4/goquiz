package entity

import (
	"sync"

	"github.com/asdine/storm"
)

var (
	DBName = "goquiz.db"

	db       *storm.DB
	openOnce sync.Once
)

func DB() *storm.DB {
	return db
}

func Open() (err error) {
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
