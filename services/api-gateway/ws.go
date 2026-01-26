package main

import (
	"log"
	"net/http"
	"ride-sharing/services/api-gateway/grpc_clients"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/proto/driver"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleRidersWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)

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

	for {
		_, message, err := conn.ReadMessage()

		if err != nil {
			log.Printf("Error reading message %v", message)
			break
		}

	}

}

func handleDriversWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)

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
		Type: "driver.cmd.register",
		Data: resp.Driver,
	}

	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Error sending message: %v", err)
		return
	}

	for {
		_, message, err := conn.ReadMessage()

		if err != nil {
			log.Printf("Error reading message %v", message)
			break
		}

	}
}
