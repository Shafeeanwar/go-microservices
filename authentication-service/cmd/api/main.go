package main

import (
	"authentication/data"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type Config struct {
	Repo   data.Repository
	Client *http.Client
}

const port = "80"

var counts int64

func main() {

	log.Println("Starting authentication service")

	//connect to DB
	conn := connectToDB()
	if conn == nil {
		log.Panic("Can't connect to Database")
	}

	//setup config
	app := Config{
		Client: &http.Client{},
	}
	app.setupRepo(conn)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")
	pgPassword := os.Getenv("PGPASSWORD")
	connectionString := strings.Replace(dsn, "PGPASSWORD", pgPassword, 1)

	for {
		connection, err := openDB(connectionString)
		if err != nil {
			log.Println("Postgress not yet ready ...")
			counts++
		} else {
			log.Println("Connected to Postgress")
			return connection
		}

		if counts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Sleeping for 2 seconds")
		time.Sleep(2 * time.Second)
		continue
	}
}

func (app *Config) setupRepo(conn *sql.DB) {
	db := data.NewPostgresRepository(conn)
	app.Repo = db
}
