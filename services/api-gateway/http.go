package main

import (
	"encoding/json"
	"log"
	"net/http"
	"ride-sharing/services/api-gateway/grpc_clients"
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

	tripService, err := grpc_clients.NewTripServiceClient()

	if err != nil {
		http.Error(w, "user id required", http.StatusInternalServerError)
		return
	}

	defer tripService.Close()

	tripprev, err := tripService.Client.PreviewTrip(r.Context(), previewTrip.ToProto())

	if err != nil {
		log.Printf("failed to preview a trip %v", err)
		http.Error(w, "failed to preview", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, contracts.APIResponse{
		Data:  tripprev,
		Error: nil,
	})
}
