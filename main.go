package main

import (
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"treffly/api"
	db "treffly/db/sqlc"
)

const (
	dbDriver      = "pgx"
	dbSource      = "postgresql://postgres@localhost:5432/postgres?sslmode=disable&password=balls"
	serverAddress = "localhost:8080"
)

func main() {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db")
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(serverAddress)
	if err != nil {
		log.Fatal(err)
	}
}
