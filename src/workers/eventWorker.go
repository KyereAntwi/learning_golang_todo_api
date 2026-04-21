package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"todoapi.com/m/src/utils"
)

type EventBody struct {
	EventType string `json:"event_type"`
}

func main() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	rabbitConn, err := utils.SetupRabbitMQ("todos_queue")
	utils.FailOnError(err, "Failed to setup RabbitMQ")
	defer rabbitConn.Close()

	db, err := sql.Open("postgres", os.Getenv("TODOS_DB_CONNECTION_STRING"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Worker connected to the database successfully")

	msgs, err := rabbitConn.Channel.Consume(rabbitConn.Queue.Name, "", false, false, false, false, nil)
	utils.FailOnError(err, "Failed to register a consumer")

	forever := make(chan struct{})

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			var eventBody EventBody
			err := json.Unmarshal(d.Body, &eventBody)
			if err != nil {
				log.Printf("Error unmarshaling event body: %v", err)
				d.Ack(false)
				continue
			}

			switch eventBody.EventType {
			case "todo_created":
				log.Println("Processing todo_created event")
			case "todo_updated":
				err := persistUpdatedTodoAuditEntry(d.Body, db)
				if err != nil {
					log.Printf("Error processing todo_updated event: %v", err)
					d.Nack(false, true)
					continue
				}
			case "todo_completed":
				log.Println("Processing todo_completed event")
			default:
				log.Printf("Unknown event type: %s", eventBody.EventType)
			}

			d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
