package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/retry"
	"ride-sharing/shared/tracing"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	TripExchange       = "trip"
	DeadLetterExchange = "DLX"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewRabbitConnection(uri string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()

	if err != nil {
		conn.Close()
		return nil, err
	}

	rmq := &RabbitMQ{conn: conn,
		Channel: ch}

	if err := rmq.setupExchangesandQueues(); err != nil {
		rmq.Close()
		return nil, err
	}

	return rmq, nil
}

func (r *RabbitMQ) setupDeadLetterQueue() error {
	err := r.Channel.ExchangeDeclare(
		DeadLetterExchange, // name
		"topic",            // type
		true,               // durable
		false,              // auto-deleted
		false,              // internal
		false,              // no-wait
		nil,                // arguments
	)

	if err != nil {
		return err
	}

	q, err := r.Channel.QueueDeclare(
		DeadLetterQueue, // name
		true,            // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		return err
	}

	err = r.Channel.QueueBind(
		q.Name,             // queue name
		"#",                // routing key
		DeadLetterExchange, // exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *RabbitMQ) setupExchangesandQueues() error {
	err := r.Channel.ExchangeDeclare(
		TripExchange, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)

	if err != nil {
		return fmt.Errorf("failed to declare exchange %v", err)
	}

	if err := r.setupDeadLetterQueue(); err != nil {
		return err
	}

	err = r.DeclareAndBindQueue(
		FindAvailableDriversQueue,
		[]string{contracts.TripEventCreated, contracts.TripEventDriverNotInterested},
		TripExchange)

	if err != nil {
		return err
	}

	if err := r.DeclareAndBindQueue(
		DriverCmdTripRequestQueue,
		[]string{contracts.DriverCmdTripRequest},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.DeclareAndBindQueue(
		DriverTripResponseQueue,
		[]string{contracts.DriverCmdTripAccept, contracts.DriverCmdTripDecline},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.DeclareAndBindQueue(
		NotifyDriverNoDriversFoundQueue,
		[]string{contracts.TripEventNoDriversFound},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.DeclareAndBindQueue(
		NotifyDriverAssignedQueue,
		[]string{contracts.TripEventDriverAssigned},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.DeclareAndBindQueue(
		PaymentTripResponseQueue,
		[]string{contracts.PaymentCmdCreateSession},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.DeclareAndBindQueue(
		NotifySessionCreatedQueue,
		[]string{contracts.PaymentEventSessionCreated},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.DeclareAndBindQueue(
		NotifyPaymentSuccessQueue,
		[]string{contracts.PaymentEventSuccess},
		TripExchange,
	); err != nil {
		return err
	}

	return nil
}

func (r *RabbitMQ) DeclareAndBindQueue(queueName string, messagesTypes []string, exchange string) error {

	args := amqp.Table{
		"x-dead-letter-queue": DeadLetterExchange,
	}

	q, err := r.Channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		args,      // arguments with dead letter exchange config
	)
	if err != nil {
		return err
	}
	for _, msg := range messagesTypes {
		err := r.Channel.QueueBind(
			q.Name,   // queue name
			msg,      // routing key
			exchange, // exchange
			false,
			nil,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *RabbitMQ) PublishMessage(ctx context.Context, routingKey string, message contracts.AmqpMessage) error {
	jsonMsg, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message")
	}
	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         jsonMsg,
	}
	return tracing.TracedPublisher(ctx, TripExchange, routingKey, msg, r.publish)

}

func (r *RabbitMQ) publish(ctx context.Context, exchange, routingKey string, msg amqp.Publishing) error {
	return r.Channel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false,
		false,
		msg,
	)
}

type MessageHandler func(context.Context, amqp.Delivery) error

func (r *RabbitMQ) ConsumeMessage(queueName string, handler MessageHandler) error {
	// Set prefetch count to 1 for fair dispatch
	// This tells RabbitMQ not to give more than one message to a service at a time.
	// The worker will only get the next message after it has acknowledged the previous one.
	err := r.Channel.Qos(
		1,     // prefetchCount: Limit to 1 unacknowledged message per consumer
		0,     // prefetchSize: No specific limit on message size
		false, // global: Apply prefetchCount to each consumer individually
	)

	if err != nil {
		return fmt.Errorf("failed to set QoS: %v", err)
	}
	msgs, err := r.Channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)

	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			//log.Printf("Received a message: %s", msg.Body)
			if err := tracing.TracedConsumer(msg, func(ctx context.Context, d amqp.Delivery) error {
				cfg := retry.DefaultConfig()
				err := retry.WithBackoff(ctx, cfg, func() error {
					return handler(ctx, d)
				})

				if err != nil {
					log.Printf("message processing failed after %d retries for message id: %s, err: %v", cfg.MaxRetries, d.MessageId, err)

					//add failure context before sending to dlq

					headers := amqp.Table{}

					if d.Headers != nil {
						headers = d.Headers
					}
					headers["x-death-reason"] = err.Error()
					headers["x-origin-exchange"] = d.Exchange
					headers["x-original-routing-key"] = d.RoutingKey
					headers["x-retry-count"] = cfg.MaxRetries
					d.Headers = headers

					_ = d.Reject(false)
					return err
				}

				if ackError := msg.Ack(false); err != nil {
					log.Printf("ERROR: Failed to Ack message: %v. Message body: %s", ackError, msg.Body)
				}
				return nil
			}); err != nil {
				log.Printf("Error processing message: %v", err)
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

	return nil
}

func (r *RabbitMQ) Close() {
	if r.conn != nil {
		r.conn.Close()
	}

	if r.Channel != nil {
		r.Channel.Close()
	}

}
