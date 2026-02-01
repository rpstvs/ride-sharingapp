package main

import (
	"context"
	"log"
	"ride-sharing/shared/messaging"

	"github.com/rabbitmq/amqp091-go"
)

type TripEventConsumer struct {
	rabbitmq *messaging.RabbitMQ
}

func NewTripEventConsumer(rbmq *messaging.RabbitMQ) *TripEventConsumer {
	return &TripEventConsumer{
		rabbitmq: rbmq,
	}
}

func (c *TripEventConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessage("hello", func(ctx context.Context, msg amqp091.Delivery) error {
		log.Println("driver received message", msg.Body)
		msg.Ack(false)
		return nil
	})
}
