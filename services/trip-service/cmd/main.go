package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/services/trip-service/internal/infrastructure/events"
	"ride-sharing/services/trip-service/internal/infrastructure/grpc"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"ride-sharing/shared/db"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"
	"ride-sharing/shared/tracing"
	"syscall"

	grpcserver "google.golang.org/grpc"
)

var GRPC_ADDR = ":9093"

func main() {

	inmemRepo := repository.NewInmemRepository()
	svc := service.NewTripService(inmemRepo)

	tracerCfg := tracing.Config{
		ServiceName:      "trip-service",
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

	mongoClient, err := db.NewMongoClient(ctx, db.NewMongoConfig())

	if err != nil {
		log.Fatalf("failed to inialize mongodb")
	}

	defer mongoClient.Disconnect(ctx)

	mongoDb := db.GetDatabase(mongoClient, db.NewMongoConfig())

	log.Println(mongoDb)
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		<-sigCh
		cancel()
	}()

	lis, err := net.Listen("tcp", GRPC_ADDR)

	if err != nil {
		log.Fatalf("failed to liste: %v", err)
	}

	rabbitMQConn, err := messaging.NewRabbitConnection(env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/"))

	if err != nil {
		log.Fatal("failed to connect to rabbitmq ")
	}

	defer rabbitMQConn.Close()

	publisher := events.NewTripEventPublisher(rabbitMQConn)

	driver_consumer := events.NewdriverConsumer(rabbitMQConn, svc)
	go driver_consumer.Listen()

	payment_consumer := events.NewpaymentConsumer(rabbitMQConn, svc)
	go payment_consumer.Listen()

	grpc_server := grpcserver.NewServer(tracing.WithTracingInterceptors()...)

	grpc.NewgRPCHandler(grpc_server, svc, publisher)

	log.Printf("Starting grpc server")

	go func() {
		if err := grpc_server.Serve(lis); err != nil {
			log.Println("failed to serve: %v", err)
			cancel()
		}
	}()

	// wait for the shutdown
	<-ctx.Done()
	grpc_server.GracefulStop()
}
