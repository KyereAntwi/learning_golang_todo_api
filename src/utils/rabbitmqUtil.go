package utils

import (
	"fmt"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConnection struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
	Queue   amqp.Queue
}

func (r *RabbitMQConnection) Close() {
	if r.Channel != nil {
		r.Channel.Close()
	}
	if r.Conn != nil {
		r.Conn.Close()
	}
}

func SetupRabbitMQ(queueName string) (*RabbitMQConnection, error) {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_CONNECTION_STRING"))

	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	} else {
		log.Println("Successfully connected to RabbitMQ")
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare a queue: %w", err)
	}

	return &RabbitMQConnection{
		Conn:    conn,
		Channel: ch,
		Queue:   q,
	}, nil
}

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
