package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand/v2"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"

	"github.com/rabbitmq/amqp091-go"
)

type TripEventConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  *Service
}

func NewTripEventConsumer(rbmq *messaging.RabbitMQ, svc *Service) *TripEventConsumer {
	return &TripEventConsumer{
		rabbitmq: rbmq,
		service:  svc,
	}
}

func (c *TripEventConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessage(messaging.FindAvailableDriversQueue, func(ctx context.Context, msg amqp091.Delivery) error {
		//log.Println("driver received message", msg.Body)

		var tripEvent contracts.AmqpMessage

		if err := json.Unmarshal(msg.Body, &tripEvent); err != nil {
			return err
		}

		var payload messaging.TripEvent

		if err := json.Unmarshal(tripEvent.Data, &payload); err != nil {
			return err
		}

		switch msg.RoutingKey {
		case contracts.TripEventCreated, contracts.TripEventDriverNotInterested:
			return c.handleFindandNotifyDrivers(ctx, payload)
		}
		log.Println("Unknown Trip Event")
		return nil
	})
}

func (c *TripEventConsumer) handleFindandNotifyDrivers(ctx context.Context, payload messaging.TripEvent) error {
	suitableDrivers := c.service.FindAvailableDrivers(payload.Trip.SelectedFare.PackageSlug)

	if len(suitableDrivers) == 0 {
		if err := c.rabbitmq.PublishMessage(ctx, contracts.TripEventNoDriversFound, contracts.AmqpMessage{
			OwnerID: payload.Trip.UserID,
		}); err != nil {

		}
		return nil
	}

	randIndex := rand.IntN(len(suitableDrivers))

	suitableDriver := suitableDrivers[randIndex]

	marshalEvent, err := json.Marshal(payload)

	if err != nil {
		return err
	}

	if err := c.rabbitmq.PublishMessage(ctx, contracts.DriverCmdTripRequest, contracts.AmqpMessage{
		OwnerID: suitableDriver,
		Data:    marshalEvent,
	}); err != nil {
		log.Printf("Error sending message")
	}

	return nil
}
