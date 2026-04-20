package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const version = "0.1.0"

type config struct {
	port             int
	env              string
	version          string
	connectionString string
}

type application struct {
	config config
	logger *log.Logger
	db     *sql.DB
}

func main() {
	var config config

	flag.IntVar(&config.port, "port", 3000, "Server port to listen on")
	flag.StringVar(&config.env, "env", "dev", "Application environment {dev|prod|staging}")
	flag.StringVar(&config.version, "version", version, "Application version")
	flag.StringVar(&config.connectionString, "db-connection-string", os.Getenv("TODOS_DB_CONNECTION_STRING"), "Database connection string")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	db, err := sql.Open("postgres", config.connectionString)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	app := &application{
		config: config,
		logger: logger,
		db:     db,
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.port),
		Handler:      app.routes(),
		ErrorLog:     logger,
		IdleTimeout:  time.Minute,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	app.logger.Printf("Database connection successful")
	app.logger.Printf("Starting server in %s mode on port %d", app.config.env, app.config.port)

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
