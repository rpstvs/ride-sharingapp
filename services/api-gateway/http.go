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

func HandleTripStart(w http.ResponseWriter, r *http.Request) {
	var tripStartRequest startTripRequest

	err := json.NewDecoder(r.Body).Decode(&tripStartRequest)

	defer r.Body.Close()

	if err != nil {
		http.Error(w, "error", http.StatusBadRequest)
		return
	}

	if tripStartRequest.UserID == "" {
		http.Error(w, "userid required", http.StatusBadRequest)
		return
	}

	if tripStartRequest.RideFareID == "" {
		http.Error(w, "ridefareid required", http.StatusBadRequest)
		return
	}

	tripService, err := grpc_clients.NewTripServiceClient()

	starttrip, err := tripService.Client.StartTrip(r.Context(), tripStartRequest.ToProto())

	if err != nil {
		http.Error(w, "error starting trip", http.StatusInternalServerError)
		return
	}

	resp := contracts.APIResponse{
		Data: starttrip,
	}

	writeJSON(w, http.StatusOK, resp)
}
