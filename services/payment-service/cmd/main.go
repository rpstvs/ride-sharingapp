package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"ride-sharing/services/payment-service/internal/infrastructure/events"
	"ride-sharing/services/payment-service/internal/infrastructure/stripe"
	"ride-sharing/services/payment-service/internal/service"
	types "ride-sharing/services/payment-service/pkg"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"
	"ride-sharing/shared/tracing"
	"syscall"
)

var GrpcAddr = env.GetString("GRPC_ADDR", ":9004")

func main() {

	rabbitMqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")

	tracerCfg := tracing.Config{
		ServiceName:      "driver-service",
		Environment:      env.GetString("ENVIRONMENT", "development"),
		ExporterEndpoint: env.GetString("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces"),
	}

	sh, err := tracing.InitTracer(tracerCfg)
	if err != nil {
		log.Fatalf("failed to inialize the tracer")
	}

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()
	defer sh(ctx)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		<-sigCh
		cancel()
	}()

	appURL := env.GetString("APP_URL", "http://localhost:3000")

	stripeCfg := &types.PaymentConfig{
		StripeSecretKey: env.GetString("STRIPE_SECRET_KEY", ""),
		SuccessURL:      env.GetString("STRIPE_SUCCESS_URL", appURL+"?payment=success"),
		CancelURL:       env.GetString("STRIPE_CANCEL_URL", appURL+"?payment=cancel"),
	}
	if stripeCfg.StripeSecretKey == "" {
		log.Fatalf("STRIPE_SECRET_KEY is not set")
		return
	}

	paymentProcessor := stripe.NewStripeClient(stripeCfg)

	service := service.NewPaymentService(paymentProcessor)

	rabbitmq, err := messaging.NewRabbitConnection(rabbitMqURI)

	if err != nil {
		log.Fatal("failed to start rabbit")
	}

	trip_consumer := events.NewTripConsumer(rabbitmq, service)

	go trip_consumer.Listen()

	defer rabbitmq.Close()
	<-ctx.Done()

}
