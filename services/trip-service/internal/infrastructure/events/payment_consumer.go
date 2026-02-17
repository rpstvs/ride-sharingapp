package events

import (
	"context"
	"encoding/json"
	"log"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"

	"github.com/rabbitmq/amqp091-go"
)

type paymentConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  domain.TripService
}

func NewpaymentConsumer(rbmq *messaging.RabbitMQ, svc domain.TripService) *paymentConsumer {
	return &paymentConsumer{
		rabbitmq: rbmq,
		service:  svc,
	}
}

func (c *paymentConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessage(messaging.NotifyPaymentSuccessQueue, func(ctx context.Context, msg amqp091.Delivery) error {
		log.Println("driver received message", msg.Body)

		var message contracts.AmqpMessage

		if err := json.Unmarshal(msg.Body, &message); err != nil {
			return err
		}

		var payload messaging.PaymentStatusUpdateData

		if err := json.Unmarshal(message.Data, &payload); err != nil {
			return err
		}

		return c.service.UpdateTrip(ctx, payload.TripID, "payed", nil)

	})
}
