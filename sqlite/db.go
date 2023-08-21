package sqlite

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func NewDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./mealshuffler.db")
	if err != nil {
		return nil, err
	}

	return db, nil
}
