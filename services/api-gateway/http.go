package main

import (
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

	//CALL TRIP SERVICE
	log.Println("SUCCESS")
	writeJSON(w, http.StatusCreated, contracts.APIResponse{
		Data:  "cenas",
		Error: nil,
	})
}
