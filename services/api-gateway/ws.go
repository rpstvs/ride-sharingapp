package main

import (
	"log"
	"net/http"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/util"

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

	type Driver struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		ProfilePicture string `json:"profilepicture"`
		CarPlate       string `json:"carplate"`
		PackageSlug    string `json:"packageslug"`
	}

	msg := contracts.WSMessage{
		Type: "driver.cmd.register",
		Data: Driver{
			ID:             userId,
			Name:           "Cenas",
			ProfilePicture: util.GetRandomAvatar(),
			PackageSlug:    packageSlug,
		},
	}

	if err := conn.WriteJSON(msg); err != nil {
		log.Println("Error sending message: %v", err)
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
