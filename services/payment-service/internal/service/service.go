package service

import (
	"context"
	"ride-sharing/services/payment-service/internal/domain"
	types "ride-sharing/services/payment-service/pkg"
	"time"

	"github.com/google/uuid"
)

type paymentService struct {
	paymentProcessor domain.PaymentProcessor
}

func NewPaymentService(paymentProcessor domain.PaymentProcessor) domain.Service {
	return &paymentService{
		paymentProcessor: paymentProcessor,
	}
}

func (p *paymentService) CreatePaymentSession(ctx context.Context, tripId, userId, driverId string, amount int64, currency string) (*types.PaymentIntent, error) {

	metadata := map[string]string{
		"trip_id":   tripId,
		"user_id":   userId,
		"driver_id": driverId,
	}

	sessionId, err := p.paymentProcessor.CreatePaymentSession(ctx, amount, currency, metadata)

	if err != nil {
		return nil, err
	}

	paymentIntent := &types.PaymentIntent{
		ID:              uuid.New().String(),
		TripID:          tripId,
		UserID:          userId,
		DriverID:        driverId,
		Amount:          amount,
		Currency:        currency,
		StripeSessionID: sessionId,
		CreatedAt:       time.Now(),
	}

	return paymentIntent, nil
}
