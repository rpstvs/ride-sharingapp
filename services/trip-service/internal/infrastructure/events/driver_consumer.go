package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"
	pbd "ride-sharing/shared/proto/driver"

	"github.com/rabbitmq/amqp091-go"
)

type driverConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  domain.TripService
}

func NewdriverConsumer(rbmq *messaging.RabbitMQ, svc domain.TripService) *driverConsumer {
	return &driverConsumer{
		rabbitmq: rbmq,
		service:  svc,
	}
}

func (c *driverConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessage(messaging.DriverTripResponseQueue, func(ctx context.Context, msg amqp091.Delivery) error {
		log.Println("driver received message", msg.Body)

		var message contracts.AmqpMessage

		if err := json.Unmarshal(msg.Body, &message); err != nil {
			return err
		}

		var payload messaging.DriveTripResponseData

		if err := json.Unmarshal(message.Data, &payload); err != nil {
			return err
		}

		switch msg.RoutingKey {
		case contracts.DriverCmdTripAccept:
			if err := c.handleTripAccepted(ctx, payload.TripID, payload.Driver); err != nil {
				log.Printf("failed to handle the trip accept")
				return err
			}
		case contracts.DriverCmdTripDecline:
			if err := c.handleTripDeclined(ctx, payload.TripID, payload.RiderID); err != nil {
				return err
			}
			return nil
		default:
			log.Println("Unknown Trip Event")
		}

		return nil
	})
}

func (c *driverConsumer) handleTripAccepted(ctx context.Context, tripID string, driver *pbd.Driver) error {
	trip, err := c.service.GetTripbyId(ctx, tripID)

	if err != nil {
		return nil
	}

	if trip == nil {
		return fmt.Errorf("trip not found")
	}

	if err := c.service.UpdateTrip(ctx, tripID, "accepted", driver); err != nil {
		log.Printf("failed to update trip")
		return err
	}

	trip, err = c.service.GetTripbyId(ctx, tripID)

	if err != nil {
		return nil
	}

	marshalledTrip, err := json.Marshal(trip)

	if err != nil {
		return err
	}

	if err := c.rabbitmq.PublishMessage(ctx, contracts.TripEventDriverAssigned, contracts.AmqpMessage{
		OwnerID: trip.UserID,
		Data:    marshalledTrip,
	}); err != nil {
		return err
	}

	marshalledPayload, err := json.Marshal(messaging.PaymentTripResponseData{
		TripID:   tripID,
		UserID:   trip.UserID,
		DriverID: driver.Id,
		Amount:   trip.RideFare.TotalPriceInCents,
		Currency: "USD",
	})

	if err != nil {
		return err
	}

	if err := c.rabbitmq.PublishMessage(ctx, contracts.PaymentCmdCreateSession, contracts.AmqpMessage{
		OwnerID: trip.UserID,
		Data:    marshalledPayload,
	}); err != nil {
		return nil
	}

	return nil
}

func (c *driverConsumer) handleTripDeclined(ctx context.Context, tripId string, RiderId string) error {

	trip, err := c.service.GetTripbyId(ctx, tripId)

	if err != nil {
		return err
	}

	newPayload := messaging.TripEvent{
		Trip: trip.ToProto(),
	}

	masrshalled, err := json.Marshal(newPayload)

	if err != nil {
		return err
	}

	if err := c.rabbitmq.PublishMessage(ctx, contracts.TripEventDriverNotInterested, contracts.AmqpMessage{
		OwnerID: trip.UserID,
		Data:    masrshalled,
	}); err != nil {
		return err
	}
	return nil
}
