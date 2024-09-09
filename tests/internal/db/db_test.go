package db_test

import (
	"github.com/stretchr/testify/assert"
	"project2/internal/config"
	"project2/internal/db"
	"testing"
)

func TestDb_InitClient(t *testing.T) {

	t.Run("Successful retrieval of client", func(t *testing.T) {
		client, err := db.InitClient()
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

}
func TestLoadDBConfig(t *testing.T) {
	// Helper function to set config values
	setConfig := func(host string, port int, user, password, dbname string) {
		config.Host = host
		config.Port = port
		config.User = user
		config.Password = password
		config.Dbname = dbname
	}

	t.Run("Valid configuration", func(t *testing.T) {
		setConfig("localhost", 5432, "user", "password", "testdb")

		dbConfig, err := db.LoadDBConfig()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if dbConfig.Host != "localhost" || dbConfig.Port != 5432 || dbConfig.User != "user" ||
			dbConfig.Password != "password" || dbConfig.DBName != "testdb" {
			t.Errorf("Unexpected config values: %+v", dbConfig)
		}
	})

	t.Run("Missing host", func(t *testing.T) {
		setConfig("", 5432, "user", "password", "testdb")

		_, err := db.LoadDBConfig()
		if err == nil {
			t.Fatal("Expected an error, got nil")
		}
		if err.Error() != "missing database configuration" {
			t.Errorf("Unexpected error message: %v", err)
		}
	})

	t.Run("Missing user", func(t *testing.T) {
		setConfig("localhost", 5432, "", "password", "testdb")

		_, err := db.LoadDBConfig()
		if err == nil {
			t.Fatal("Expected an error, got nil")
		}
		if err.Error() != "missing database configuration" {
			t.Errorf("Unexpected error message: %v", err)
		}
	})

	t.Run("Missing password", func(t *testing.T) {
		setConfig("localhost", 5432, "user", "", "testdb")

		_, err := db.LoadDBConfig()
		if err == nil {
			t.Fatal("Expected an error, got nil")
		}
		if err.Error() != "missing database configuration" {
			t.Errorf("Unexpected error message: %v", err)
		}
	})

	t.Run("Missing dbname", func(t *testing.T) {
		setConfig("localhost", 5432, "user", "password", "")

		_, err := db.LoadDBConfig()
		if err == nil {
			t.Fatal("Expected an error, got nil")
		}
		if err.Error() != "missing database configuration" {
			t.Errorf("Unexpected error message: %v", err)
		}
	})
}
