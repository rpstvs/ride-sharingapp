package tripservice

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
	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"
	"syscall"
	"time"

	grpcserver "google.golang.org/grpc"
)

var GRPC_ADDR = ":9093"

func main() {

	inmemRepo := repository.NewInmemRepository()
	svc := service.NewTripService(inmemRepo)

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

	rabbitMQConn, err := messaging.NewRabbitConnection(env.GetString(env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")))

	if err != nil {
		log.Fatal("failed to connect to rabbitmq ")
	}

	defer rabbitMQConn.Close()

	publisher := events.NewTripEventPublisher(rabbitMQConn)

	grpc_server := grpcserver.NewServer()

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
