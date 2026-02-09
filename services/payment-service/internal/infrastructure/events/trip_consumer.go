package events

import (
	"context"
	"encoding/json"
	"log"
	"ride-sharing/services/payment-service/internal/domain"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"

	"github.com/rabbitmq/amqp091-go"
)

type TripConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  domain.Service
}

func NewTripConsumer(rbmq *messaging.RabbitMQ, svc domain.Service) *TripConsumer {
	return &TripConsumer{
		rabbitmq: rbmq,
		service:  svc,
	}
}

func (c *TripConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessage(messaging.PaymentTripResponseQueue, func(ctx context.Context, msg amqp091.Delivery) error {

		var message contracts.AmqpMessage

		if err := json.Unmarshal(msg.Body, &message); err != nil {
			return err
		}

		var payload messaging.PaymentTripResponseData

		if err := json.Unmarshal(message.Data, &payload); err != nil {
			return err
		}

		switch msg.RoutingKey {
		case contracts.PaymentCmdCreateSession:
			return c.handleTripAccepted(ctx, payload)
		}
		log.Println("Unknown Trip Event")
		return nil
	})
}

func (c *TripConsumer) handleTripAccepted(ctx context.Context, payload messaging.PaymentTripResponseData) error {

	paymentIntent, err := c.service.CreatePaymentSession(ctx, payload.TripID, payload.UserID, payload.DriverID, int64(payload.Amount), payload.Currency)

	if err != nil {
		return err
	}

	paymentPayload := messaging.PaymentEventSessionCreatedData{
		TripID:    payload.TripID,
		SessionID: paymentIntent.ID,
		Currency:  payload.Currency,
		Amount:    float64(paymentIntent.Amount) / 100.0,
	}

	payloadBytes, err := json.Marshal(&paymentPayload)

	if err != nil {
		return err
	}

	if err := c.rabbitmq.PublishMessage(ctx, contracts.PaymentEventSessionCreated, contracts.AmqpMessage{
		OwnerID: payload.UserID,
		Data:    payloadBytes,
	}); err != nil {
		return err
	}

	return nil
}
