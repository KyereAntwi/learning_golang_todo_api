package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"todoapi.com/m/src/traceConfigurations"
	"todoapi.com/m/src/utils"
	. "todoapi.com/m/src/utils"

	_ "todoapi.com/m/docs"
)

type config struct {
	port             int
	env              string
	version          string
	connectionString string
	jwtSecret        string
	appHost          string
}

type application struct {
	config     config
	logger     *log.Logger
	db         *sql.DB
	rabbitConn *RabbitMQConnection
	jwtManager *JWTManager
}

// @title TODO Learning with Go API
// @version 0.30
// @description This is a TODO learning API built with Go.
// @host localhost:3000
// @BasePath /api/v1
func main() {

	// initialize OpenTelemetry tracer
	tracer, err := traceConfigurations.InitTracer()
	if err != nil {
		log.Fatalf("Error initializing OpenTelemetry tracer: %v", err)
	}
	defer func() {
		if err := tracer.Shutdown(context.Background()); err != nil {
			log.Fatalf("Error shutting down OpenTelemetry tracer: %v", err)
		}
	}()

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
	flag.StringVar(&config.jwtSecret, "jwt-secret", os.Getenv("JWT_SECRET"), "JWT secret key")
	flag.StringVar(&config.appHost, "app-host", os.Getenv("APP_HOST"), "Application host")
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
		jwtManager: NewJWTManager(&config),
	}

	middleware := NewLoggerMiddleware(logger)
	authMiddleware := NewAuthMiddleware(app.jwtManager)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.port),
		Handler:      middleware(authMiddleware(app.routes())),
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

	// Graceful shutdown logic with context and signal handling can be added here
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		log.Println("Shutting down server...")

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Printf("Error shutting down server: %v", err)
		} else {
			log.Println("Server gracefully stopped")
		}
	}()
	wg.Wait()
}
