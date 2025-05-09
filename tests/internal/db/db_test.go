package db_test

import (
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"

	"project2/internal/config"
	"project2/internal/db"
)

func TestInitClient_Success(t *testing.T) {
	// Create a new sqlmock instance with ping monitoring enabled
	mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer mockDB.Close()

	// Replace the actual sql.Open function with a mock
	initialOpen := db.CreateConn
	db.CreateConn = func(driverName, dataSourceName string) (*sql.DB, error) {
		if driverName == "postgres" {
			return mockDB, nil
		}
		return nil, errors.New("unexpected driver name")
	}
	defer func() {
		db.CreateConn = initialOpen
	}()

	// Mock the Ping method
	mock.ExpectPing().WillReturnError(nil)

	// Ensure that InitClient works with the mock
	dbInstance, err := db.PostgresInitClient()
	assert.NoError(t, err)
	assert.NotNil(t, dbInstance)

	// Ensure that all mocked expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}

func TestInitClient_PingError(t *testing.T) {
	// Create a new sqlmock instance with ping monitoring enabled
	mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer mockDB.Close()

	// Replace the actual sql.Open function with a mock
	initialOpen := db.CreateConn
	db.CreateConn = func(driverName, dataSourceName string) (*sql.DB, error) {
		if driverName == "postgres" {
			return mockDB, nil
		}
		return nil, errors.New("unexpected driver name")
	}
	defer func() {
		db.CreateConn = initialOpen
	}()

	// Mock the Ping method
	mock.ExpectPing().WillReturnError(errors.New("ping errs"))

	// Ensure that InitClient works with the mock
	_, err = db.PostgresInitClient()
	assert.Error(t, err)

	// Ensure that all mocked expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}

}

func TestInitClient_ConnectionError(t *testing.T) {
	// Create a new sqlmock instance
	_, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	// Replace the actual sql.Open function with a mock
	initialOpen := db.CreateConn
	db.CreateConn = func(driverName, dataSourceName string) (*sql.DB, error) {
		return nil, errors.New("failed to open connection")
	}
	defer func() {
		db.CreateConn = initialOpen
	}()

	dbInstance, err := db.PostgresInitClient()
	assert.Error(t, err)
	assert.Nil(t, dbInstance)
}

func TestInitClient_LoadingConfigError(t *testing.T) {
	// Set a mock function to avoid actual DB connection
	initialConn := db.CreateConn
	db.CreateConn = func(driverName, dataSourceName string) (*sql.DB, error) {
		return &sql.DB{}, nil
	}
	defer func() {
		db.CreateConn = initialConn
	}()

	// Helper function to set config values
	setConfig := func(host string, port int, user, password, dbname string) {
		config.Host = host
		config.Port = port
		config.User = user
		config.Password = password
		config.Dbname = dbname
	}

	setConfig("", 5432, "user", "password", "testdb")
	dbInstance, err := db.PostgresInitClient()

	assert.Error(t, err)
	assert.Nil(t, dbInstance)
}

func TestSingletonPattern(t *testing.T) {
	// Create a new sqlmock instance
	mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer mockDB.Close()

	// Replace the actual sql.Open function with a mock
	initialOpen := db.CreateConn
	db.CreateConn = func(driverName, dataSourceName string) (*sql.DB, error) {
		if driverName == "postgres" {
			return mockDB, nil
		}
		return nil, errors.New("unexpected driver name")
	}
	defer func() {
		db.CreateConn = initialOpen
	}()

	// Mock the Ping method
	mock.ExpectPing().WillReturnError(nil)

	// Call InitClient multiple times
	client1, err1 := db.PostgresInitClient()
	if err1 != nil {
		t.Fatalf("Error initializing client: %v", err1)
	}

	client2, err2 := db.PostgresInitClient()
	if err2 != nil {
		t.Fatalf("Error initializing client: %v", err2)
	}

	client3, err3 := db.PostgresInitClient()
	if err3 != nil {
		t.Fatalf("Error initializing client: %v", err3)
	}

	// Assert that all clients are the same instance
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)

	assert.Equal(t, client1, client2, "client1 and client2 should be the same instance")
	assert.Equal(t, client1, client3, "client1 and client3 should be the same instance")
	assert.Equal(t, client2, client3, "client2 and client3 should be the same instance")
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
			t.Fatalf("Expected no errs, got %v", err)
		}

		if dbConfig.Host != config.Host || dbConfig.Port != config.Port || dbConfig.User != config.User ||
			dbConfig.Password != config.Password || dbConfig.DBName != config.Dbname {
			t.Errorf("Unexpected config values: %+v", dbConfig)
		}
	})

	t.Run("Missing host", func(t *testing.T) {
		setConfig("", 5432, "user", "password", "testdb")

		_, err := db.LoadDBConfig()
		if err == nil {
			t.Fatal("Expected an errs, got nil")
		}
		if err.Error() != "missing database configuration" {
			t.Errorf("Unexpected errs message: %v", err)
		}
	})

	t.Run("Missing user", func(t *testing.T) {
		setConfig("localhost", 5432, "", "password", "testdb")

		_, err := db.LoadDBConfig()
		if err == nil {
			t.Fatal("Expected an errs, got nil")
		}
		if err.Error() != "missing database configuration" {
			t.Errorf("Unexpected errs message: %v", err)
		}
	})

	t.Run("Missing password", func(t *testing.T) {
		setConfig("localhost", 5432, "user", "", "testdb")

		_, err := db.LoadDBConfig()
		if err == nil {
			t.Fatal("Expected an errs, got nil")
		}
		if err.Error() != "missing database configuration" {
			t.Errorf("Unexpected errs message: %v", err)
		}
	})

	t.Run("Missing dbname", func(t *testing.T) {
		setConfig("localhost", 5432, "user", "password", "")

		_, err := db.LoadDBConfig()
		if err == nil {
			t.Fatal("Expected an errs, got nil")
		}
		if err.Error() != "missing database configuration" {
			t.Errorf("Unexpected errs message: %v", err)
		}
	})
}
