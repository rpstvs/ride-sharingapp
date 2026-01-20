package main

import (
	"net/http"

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

}

func handleDriversWebsocket(w http.ResponseWriter, r *http.Request) {

}
