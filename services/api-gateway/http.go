package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"ride-sharing/services/api-gateway/grpc_clients"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"
	"ride-sharing/shared/tracing"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
)

var tracer = tracing.GetTracingInstance("api-gateway")

func HandleTripPreview(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "HandleTripPreview")
	defer span.End()

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

	tripprev, err := tripService.Client.PreviewTrip(ctx, previewTrip.ToProto())

	if err != nil {
		log.Printf("failed to preview a trip %v", err)
		http.Error(w, "failed to preview", http.StatusInternalServerError)
		return
	}

	log.Printf("Fares: %v", tripprev.RideFares)

	resp := contracts.APIResponse{
		Data: tripprev,
	}

	writeJSON(w, http.StatusCreated, resp)
}

func HandleTripStart(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "HandleTripStart")
	defer span.End()
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

	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	defer tripService.Close()

	starttrip, err := tripService.Client.StartTrip(ctx, tripStartRequest.ToProto())

	if err != nil {
		http.Error(w, "error starting trip", http.StatusInternalServerError)
		return
	}

	resp := contracts.APIResponse{
		Data: starttrip,
	}

	writeJSON(w, http.StatusOK, resp)
}

func handleStripeWebHook(w http.ResponseWriter, r *http.Request, rabbitmq *messaging.RabbitMQ) {
	ctx, span := tracer.Start(r.Context(), "handleStripeWebHook")
	defer span.End()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	webhookKey := env.GetString("STRIPE_WEBHOOK_KEY", "")

	if webhookKey == "" {
		log.Println("stripe webhook key is required")
		return
	}

	event, err := webhook.ConstructEventWithOptions(
		body,
		r.Header.Get("Stripe-Signature"),
		webhookKey,
		webhook.ConstructEventOptions{
			IgnoreAPIVersionMismatch: true,
		},
	)

	if err != nil {
		log.Printf("Error verifying webhook signature %v", err)
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}

	log.Printf("received event from stripe %v", event)

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession

		err := json.Unmarshal(event.Data.Raw, &session)

		if err != nil {
			log.Printf("error parsing webhook json %v ", err)
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		payload := messaging.PaymentStatusUpdateData{
			TripID:   session.Metadata["trip_id"],
			UserID:   session.Metadata["user_id"],
			DriverID: session.Metadata["driver_id"],
		}

		payloadBytes, err := json.Marshal(payload)

		if err != nil {
			log.Printf("failed to marshal payload")
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		if err := rabbitmq.PublishMessage(ctx, contracts.PaymentEventSuccess, contracts.AmqpMessage{
			OwnerID: session.Metadata["user_id"],
			Data:    payloadBytes,
		}); err != nil {

		}

	}
}
