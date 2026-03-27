package database

import (
	"database/sql"
	"github.com/sachinchandra/goflash/internal/config"
	"log"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() {
	connStr := config.GetEnv("DB_URL", "postgres://user:password@localhost:5432/goflash?sslmode=disable")
	
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to open DB: ", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to ping DB: ", err)
	}
	log.Println("Successfully connected to PostgreSQL!")
}