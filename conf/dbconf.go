package conf

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// type DBConn interface {
// 	Query(query string, args ...interface{}) (*sql.Rows, error)
// 	Prepare(query string) (*sql.Stmt, error)
// }

func (dbConfig DBConfig) DbConn() *sql.DB {
	// create the data source name from db config
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	return db
}
