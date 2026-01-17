package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"ride-sharing/shared/contracts"
)

func HandleTripPreview(w http.ResponseWriter, r *http.Request) {

	var previewTrip PreviewTripRequest

	if err := json.NewDecoder(r.Body).Decode(&previewTrip); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	//Validation

	if previewTrip.UserID == "" {
		http.Error(w, "user id required", http.StatusBadRequest)
		return
	}

	jsonBody, _ := json.Marshal(previewTrip)
	reader := bytes.NewReader(jsonBody)

	//CALL TRIP SERVICE

	resp, err := http.Post("http://trip-service/trip/preview", "application/json", reader)

	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	log.Println("SUCCESS")
	writeJSON(w, http.StatusCreated, contracts.APIResponse{
		Data:  resp.Body,
		Error: nil,
	})
}
