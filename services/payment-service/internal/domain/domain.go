package domain

import (
	"context"
	types "ride-sharing/services/payment-service/pkg"
)

type Service interface {
	CreatePaymentSession(ctx context.Context, tripId, userId, driverId string, amount int64, currency string) (*types.PaymentIntent, error)
}

type PaymentProcessor interface {
	CreatePaymentSession(ctx context.Context, amount int64, currency string, metadata map[string]string) (string, error)
	//GetSessionStatus(ctx context.Context, sessionId string) (*types.PaymentStatus, error)
}
