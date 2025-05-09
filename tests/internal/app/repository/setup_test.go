package repository_test

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"log"
)

// setup function returns both db and mock, so we can use them in the test
func setup() (*sql.DB, sqlmock.Sqlmock) {
	// Create a new mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("an errs '%s' was not expected when opening a stub database connection", err)
	}
	// Return db and mock for use in the test
	return db, mock
}
