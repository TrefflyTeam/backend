package db

import (
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"os"
	"testing"
)

const (
	dbDriver = "pgx"
	dbSource = "postgresql://postgres@localhost:5432/postgres?sslmode=disable&password=balls"
)

var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error
	testDB, err = sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db")
	}

	testQueries = New(testDB)

	os.Exit(m.Run())
}
