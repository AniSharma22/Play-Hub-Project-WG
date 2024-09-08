package db

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	_ "github.com/lib/pq" // Postgres driver
)

var (
	dbInstance *sql.DB
	once       sync.Once
)

// DBConfig holds the database configuration
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// LoadDBConfig loads database configuration from .env file
func LoadDBConfig() (*DBConfig, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	return &DBConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     port,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
	}, nil
}

// InitClient initializes a PostgresSQL client as a singleton with a connection pool
func InitClient() (*sql.DB, error) {
	var err error

	// Ensure that only one instance of the database is created
	once.Do(func() {
		var config *DBConfig
		config, err = LoadDBConfig()
		if err != nil {
			return
		}

		psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			config.Host, config.Port, config.User, config.Password, config.DBName)

		dbInstance, err = sql.Open("postgres", psqlInfo)
		if err != nil {
			log.Fatalf("Failed to connect to PostgreSQL: %v", err)
		}

		// Set connection pool limits
		dbInstance.SetMaxOpenConns(100)                // Maximum number of open connections to the database
		dbInstance.SetMaxIdleConns(50)                 // Maximum number of idle connections in the pool
		dbInstance.SetConnMaxLifetime(5 * time.Minute) // Connections will close and get replaced after 5 minutes

		// Check the connection
		err = dbInstance.Ping()
		if err != nil {
			log.Fatalf("Failed to ping PostgreSQL: %v", err)
		}
	})

	return dbInstance, err
}
