package main

import (
	"encoding/json"
	"log"
	"net/http"
	"ride-sharing/services/api-gateway/grpc_clients"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"
	"ride-sharing/shared/proto/driver"
)

var (
	connManager = messaging.NewConnectionManager()
)

func handleRidersWebsocket(w http.ResponseWriter, r *http.Request, rb *messaging.RabbitMQ) {
	conn, err := connManager.Upgrade(w, r)

	if err != nil {
		ErrorJson(w, http.StatusInternalServerError, "websocket upgrade fail")
		return
	}

	defer conn.Close()
	userId := r.URL.Query().Get("userID")

	if userId == "" {
		ErrorJson(w, http.StatusBadRequest, "userid is required")
		return
	}

	connManager.Add(conn, userId)
	defer connManager.Remove(userId)

	queues := []string{messaging.NotifyDriverNoDriversFoundQueue,
		messaging.NotifyDriverAssignedQueue}

	for _, q := range queues {
		consumer := messaging.NewQueueConsumer(rb, connManager, q)

		if err := consumer.Start(); err != nil {
			log.Printf("Failed to start consumer for queue: %s: err: %v", q, err)
		}
	}

	for {
		_, message, err := conn.ReadMessage()

		if err != nil {
			log.Printf("Error reading message %v", message)
			break
		}

	}

}

func handleDriversWebsocket(w http.ResponseWriter, r *http.Request, rb *messaging.RabbitMQ) {
	conn, err := connManager.Upgrade(w, r)

	if err != nil {
		ErrorJson(w, http.StatusInternalServerError, "websocket upgrade fail")
		return
	}

	defer conn.Close()

	userId := r.URL.Query().Get("userID")

	if userId == "" {
		ErrorJson(w, http.StatusBadRequest, "userid is required")
		return
	}

	connManager.Add(conn, userId)

	packageSlug := r.URL.Query().Get("packageSlug")

	if packageSlug == "" {
		log.Println("no package slug provided")
		return
	}

	driverService, err := grpc_clients.NewDriverServiceClient()

	if err != nil {
		log.Println("no package slug provided")
		return
	}

	defer func() {
		connManager.Remove(userId)
		driverService.Client.UnRegisterDriver(r.Context(), &driver.RegisterDriverRequest{
			ID:          userId,
			PackageSlug: packageSlug,
		})
		driverService.Close()

	}()

	grpcReq := &driver.RegisterDriverRequest{
		ID:          userId,
		PackageSlug: packageSlug,
	}

	resp, err := driverService.Client.RegisterDriver(r.Context(), grpcReq)

	if err != nil {
		log.Println("no package slug provided")
		return
	}

	msg := contracts.WSMessage{
		Type: contracts.DriverCmdRegister,
		Data: resp.Driver,
	}

	if err := connManager.SendMessage(userId, msg); err != nil {
		log.Printf("Error sending message: %v", err)
		return
	}

	queues := []string{messaging.DriverCmdTripRequestQueue}

	for _, q := range queues {
		consumer := messaging.NewQueueConsumer(rb, connManager, q)

		if err := consumer.Start(); err != nil {
			log.Printf("Failed to start consumer for queue: %s: err: %v", q, err)
		}
	}

	for {
		_, message, err := conn.ReadMessage()

		if err != nil {
			log.Printf("Error reading message %v", message)
			break
		}

		type DriverMessage struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		var driverMessage DriverMessage

		if err := json.Unmarshal(message, &driverMessage); err != nil {
			log.Printf("Error unmarsahing drivermessage")
			continue
		}

		switch driverMessage.Type {
		case contracts.DriverCmdLocation:
			continue
		case contracts.DriverCmdTripAccept, contracts.DriverCmdTripDecline:
			if err := rb.PublishMessage(ctx, driverMessage.Type, contracts.AmqpMessage{
				OwnerID: userId,
				Data:    driverMessage.Data,
			}); err != nil {

			}

		default:
			log.Printf("not supported")
		}
	}
}
