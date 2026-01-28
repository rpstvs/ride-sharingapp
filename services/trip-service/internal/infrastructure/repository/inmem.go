package repository

import (
	"context"
	"fmt"
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

func (m *InMemRepository) SaveRideFare(ctx context.Context, fare *domain.RideFareModel) error {
	m.rideFares[fare.ID.Hex()] = fare
	return nil
}

func (m *InMemRepository) GetRideFareByID(ctx context.Context, id string) (*domain.RideFareModel, error) {
	fare, ok := m.rideFares[id]

	if !ok {
		return nil, fmt.Errorf("fare does not exist")
	}

	return fare, nil
}
