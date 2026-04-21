package repositories

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"todoapi.com/m/src/domain"
	"todoapi.com/m/src/utils"
)

type PublisherService struct {
	rabbitConn *utils.RabbitMQConnection
}

func NewPublisherService(rabbitConn *utils.RabbitMQConnection) *PublisherService {
	return &PublisherService{
		rabbitConn: rabbitConn,
	}
}

func (p *PublisherService) PublishTodoCreated(todo domain.TodoCreatedEvent) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body, err := json.Marshal(todo)
	if err != nil {
		return err
	}

	defer p.rabbitConn.Close()

	err = p.rabbitConn.Channel.PublishWithContext(ctx, "", p.rabbitConn.Queue.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})

	if err != nil {
		return err
	}

	return nil
}

func (p *PublisherService) PublishTodoUpdated(todo domain.TodoUpdatedEvent) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body, err := json.Marshal(todo)
	if err != nil {
		return err
	}

	defer p.rabbitConn.Close()

	err = p.rabbitConn.Channel.PublishWithContext(ctx, "", p.rabbitConn.Queue.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})

	if err != nil {
		return err
	}

	return nil
}

func (p *PublisherService) PublishTodoCompleted(todo domain.TodoCompletedEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body, err := json.Marshal(todo)
	if err != nil {
		return err
	}

	defer p.rabbitConn.Close()

	err = p.rabbitConn.Channel.PublishWithContext(ctx, "", p.rabbitConn.Queue.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})

	if err != nil {
		return err
	}

	return nil
}
