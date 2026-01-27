package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"
	"syscall"
	"time"

	grpcserver "google.golang.org/grpc"
)

var GRPC_ADDR = ":9092"

func main() {
	svc := NewService()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

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

	grpc_server := grpcserver.NewServer()

	rabbitMQConn, err := messaging.NewRabbitConnection(env.GetString(env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")))

	if err != nil {
		log.Fatal("failed to connect to rabbitmq ")
	}

	defer rabbitMQConn.Close()

	NewgRPCHandler(grpc_server, svc)

	log.Printf("Starting grpc server")

	go func() {
		if err := grpc_server.Serve(lis); err != nil {
			log.Printf("failed to serve: %v", err)
			cancel()
		}
	}()

	// wait for the shutdown
	<-ctx.Done()
	grpc_server.GracefulStop()
}
