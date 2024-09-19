package db

import (
	"database/sql"
	"fmt"
	"log"
	"project2/internal/config"
	"sync"
	"time"

	_ "github.com/lib/pq" // Postgres driver
)

var (
	dbInstance *sql.DB
	once       sync.Once
	CreateConn = sql.Open
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
	if config.Host == "" || config.User == "" || config.Password == "" || config.Dbname == "" {
		return nil, fmt.Errorf("missing database configuration")
	}

	return &DBConfig{
		Host:     config.Host,
		Port:     config.Port,
		User:     config.User,
		Password: config.Password,
		DBName:   config.Dbname,
	}, nil
}

// InitClient initializes a PostgresSQL client as a singleton with a connection pool
func PostgresInitClient() (*sql.DB, error) {
	var err error

	once.Do(func() {
		// Load configuration
		var cfg *DBConfig
		cfg, err = LoadDBConfig()
		if err != nil {
			err = fmt.Errorf("failed to load DB config: %v", err)
			return
		}

		// Build connection string
		psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

		// Open connection to Postgres
		dbInstance, err = CreateConn("postgres", psqlInfo)
		if err != nil {
			err = fmt.Errorf("failed to open PostgreSQL connection: %v", err)
			return
		}

		// Set connection pool limits
		dbInstance.SetMaxOpenConns(100)                // Maximum number of open connections to the database
		dbInstance.SetMaxIdleConns(50)                 // Maximum number of idle connections in the pool
		dbInstance.SetConnMaxLifetime(5 * time.Minute) // Connections will close and get replaced after 5 minutes

		//Check the connection
		err = dbInstance.Ping()
		if err != nil {
			err = fmt.Errorf("failed to ping PostgreSQL: %v", err)
			return
		}

		log.Println("Successfully connected to PostgresSQL")
	})

	// Return dbInstance and any errs that occurred during initialization
	return dbInstance, err
}
