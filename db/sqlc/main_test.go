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
	dbSource = "postgresql://postgres@localhost:5432/postgres?sslmode=disable&password=secret"
)

var testQueries *Queries

func TestMain(m *testing.M) {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db")
	}

	testQueries = New(conn)

	os.Exit(m.Run())
}
