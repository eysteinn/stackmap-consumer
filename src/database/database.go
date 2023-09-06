package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var (
	db *sql.DB = nil
)

func setup() (*sql.DB, error) {

	Host := os.Getenv("PSQL_HOST")
	User := os.Getenv("PSQL_USER")
	DB := os.Getenv("PSQL_DB")
	Pass := os.Getenv("PSQL_PASS")
	Port := os.Getenv("PSQL_PORT")

	if Host == "" {
		Host = "postgresql.default.svc.cluster.local"
	}
	if User == "" {
		User = "postgres"
	}
	if DB == "" {
		DB = "postgres"
	}

	if Host == "" || User == "" || DB == "" || Pass == "" {
		return nil, errors.New("Unable to grap credentials for PSQL")
	}
	if Port == "" {
		Port = "5432"
	}

	conn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		Host, User, Pass, DB, Port)

	return sql.Open("postgres", conn)
}

func GetDB() (*sql.DB, error) {
	if db == nil {
		tmp, err := setup()
		if err != nil {
			return nil, err
		}
		db = tmp
	}
	return db, nil
}
