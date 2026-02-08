package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"
)

var (
	httpAddr    = env.GetString("HTTP_ADDR", ":8081")
	rabbitMqURI = env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
)

func main() {
	log.Println("Starting API Gateway")

	mux := http.NewServeMux()

	// RabbitMQ connection
	rabbitmq, err := messaging.NewRabbitConnection(rabbitMqURI)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitmq.Close()

	log.Println("Starting RabbitMQ connection")

	mux.HandleFunc("POST /trip/preview", EnableCors(HandleTripPreview))
	mux.HandleFunc("POST /trip/start", EnableCors(HandleTripStart))
	mux.HandleFunc("/ws/drivers", func(w http.ResponseWriter, r *http.Request) {
		handleDriversWebsocket(w, r, rabbitmq)
	})
	mux.HandleFunc("/ws/riders", func(w http.ResponseWriter, r *http.Request) {
		handleRidersWebsocket(w, r, rabbitmq)
	})

	server := &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Printf("Error starting server: %v", err)
	case sig := <-shutdown:
		log.Printf("server is shutting down: %v", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Println("couldnt gracefully shutdown")
		server.Close()
	}
}
