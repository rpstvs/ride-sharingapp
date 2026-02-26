package repository

import (
	"context"
	"fmt"
	"ride-sharing/services/trip-service/internal/domain"
	pbd "ride-sharing/shared/proto/driver"
	pb "ride-sharing/shared/proto/trip"
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

func (m *InMemRepository) GetTripbyId(ctx context.Context, tripId string) (*domain.TripModel, error) {
	trip, exists := m.trips[tripId]

	if !exists {
		return nil, fmt.Errorf("trip not found")
	}
	return trip, nil
}
func (m *InMemRepository) UpdateTrip(ctx context.Context, tripid string, status string, driver *pbd.Driver) error {
	trip, exists := m.trips[tripid]

	if !exists {
		return fmt.Errorf("trip not found")
	}

	trip.Status = status

	if driver != nil {
		trip.Driver = &pb.TripDriver{
			Id:             driver.Id,
			Name:           driver.Name,
			ProfilePicture: driver.ProfilePicture,
			CarPlate:       driver.CarPlate,
		}
	}

	return nil

}
