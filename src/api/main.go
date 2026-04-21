package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"todoapi.com/m/src/utils"
	. "todoapi.com/m/src/utils"
)

type config struct {
	port             int
	env              string
	version          string
	connectionString string
}

type application struct {
	config     config
	logger     *log.Logger
	db         *sql.DB
	rabbitConn *RabbitMQConnection
}

func main() {
	var config config

	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("Error converting PORT to int: %v", err)
	}

	flag.IntVar(&config.port, "port", port, "Server port to listen on")
	flag.StringVar(&config.env, "env", os.Getenv("APP_ENV"), "Application environment {dev|prod|staging}")
	flag.StringVar(&config.version, "version", os.Getenv("VERSION"), "Application version")
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

	rabbitConn, err := utils.SetupRabbitMQ("todos_queue")
	utils.FailOnError(err, "Failed to setup RabbitMQ")
	defer rabbitConn.Close()

	app := &application{
		config:     config,
		logger:     logger,
		db:         db,
		rabbitConn: rabbitConn,
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
