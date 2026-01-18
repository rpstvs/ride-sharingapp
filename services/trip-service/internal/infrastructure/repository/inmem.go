package repository

import (
	"context"
	"ride-sharing/services/trip-service/internal/domain"
)

type InMemRepository struct {
	trips     map[string]*domain.TripModel
	rideFares map[string]*domain.RideFareModel
}

func NewInmemRepository() *InMemRepository {
	return &InMemRepository{
		trips:     make(map[string]*domain.TripModel),
		rideFares: make(map[string]*domain.RideFareModel),
	}
}

func (m *InMemRepository) CreateTrip(ctx context.Context, trip *domain.TripModel) (*domain.TripModel, error) {

	m.trips[trip.ID.Hex()] = trip
	return trip, nil
}
