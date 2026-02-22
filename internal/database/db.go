package database

import(
	"database/sql"
	"fmt"
	"log"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB(){
	var err error

	connStr := "postgres://user:password@localhost:5432/goflash?sslmode=disable"

	DB, err=sql.Open("postgres",connStr);

	if err != nil{
		log.Fatal("Failed to open database connection", err)
	}
	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to ping DB: ", err)
	}
	fmt.Println("Successfully connected to PostgreSQL!")
}