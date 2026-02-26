package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"ride-sharing/services/trip-service/internal/domain"
	tripTypes "ride-sharing/services/trip-service/pkg/types"
	pbd "ride-sharing/shared/proto/driver"
	"ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TripService struct {
	Repository domain.TripRepository
}

func NewTripService(repo domain.TripRepository) *TripService {
	return &TripService{
		Repository: repo,
	}
}

func (s *TripService) CreateTrip(ctx context.Context, fare *domain.RideFareModel) (*domain.TripModel, error) {
	t := &domain.TripModel{
		ID:       primitive.NewObjectID(),
		UserID:   fare.UserID,
		Status:   "",
		RideFare: fare,
		Driver:   &trip.TripDriver{},
	}

	trip, err := s.Repository.CreateTrip(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("error creating trip")
	}

	return trip, nil
}

func (s *TripService) GetRoute(ctx context.Context, pickup, dest types.Coordinate) (*tripTypes.OsmrApiResponse, error) {
	url := fmt.Sprintf("http://router.project-osrm.org/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson",
		pickup.Longitude,
		pickup.Latitude,
		dest.Longitude,
		dest.Latitude)

	resp, err := http.Get(url)

	if err != nil {
		return nil, fmt.Errorf("error requesting route api")
	}

	defer resp.Body.Close()
	var route tripTypes.OsmrApiResponse

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("error decoding route from response")
	}

	err = json.Unmarshal(body, &route)

	if err != nil {
		return nil, fmt.Errorf("error decoding route from response")
	}

	return &route, nil
}

func (s *TripService) EstimatePackagesPricesWithRoute(route *tripTypes.OsmrApiResponse) []*domain.RideFareModel {
	baseFares := getBaseFares()

	estimatedFares := make([]*domain.RideFareModel, len(baseFares))

	for i, f := range baseFares {
		estimatedFares[i] = s.estimateFareOut(f, route)
	}
	return estimatedFares
}

func (s *TripService) GenerateTripFares(ctx context.Context, fares []*domain.RideFareModel, userid string, route *tripTypes.OsmrApiResponse) ([]*domain.RideFareModel, error) {
	ridefares := make([]*domain.RideFareModel, len(fares))

	for i, f := range fares {
		id := primitive.NewObjectID()
		fare := &domain.RideFareModel{
			UserID:            userid,
			ID:                id,
			TotalPriceInCents: f.TotalPriceInCents,
			PackageSlug:       f.PackageSlug,
			Route:             route,
		}

		if err := s.Repository.SaveRideFare(ctx, fare); err != nil {
			return nil, fmt.Errorf("failed to save trip fare")
		}

		ridefares[i] = fare
	}
	fmt.Println(ridefares)
	return ridefares, nil
}

func (s *TripService) estimateFareOut(fare *domain.RideFareModel, route *tripTypes.OsmrApiResponse) *domain.RideFareModel {
	// distance , time and car price
	pricingCfg := tripTypes.DefaultPricingConfig()
	carPackagePrice := fare.TotalPriceInCents
	distance := route.Routes[0].Distance
	duration := route.Routes[0].Duration

	distanceFare := distance * pricingCfg.PricePerUnitOfDistance
	timeFare := duration * pricingCfg.PricingPerMinute

	totalPrice := carPackagePrice + distanceFare + timeFare

	return &domain.RideFareModel{
		TotalPriceInCents: totalPrice,
		PackageSlug:       fare.PackageSlug,
	}
}

func (s *TripService) GetAndValidateFare(ctx context.Context, fareId, Userid string) (*domain.RideFareModel, error) {
	fare, err := s.Repository.GetRideFareByID(ctx, fareId)

	if err != nil {
		log.Printf("fares not found %v", err)
		return nil, err
	}

	if fare == nil {
		log.Printf("fares not found %v", err)
		return nil, fmt.Errorf("fare does not exist")
	}

	if fare.UserID != Userid {
		return nil, fmt.Errorf("userid and fareid mismatch")
	}

	return fare, nil
}

func getBaseFares() []*domain.RideFareModel {
	return []*domain.RideFareModel{
		{
			PackageSlug:       "suv",
			TotalPriceInCents: 200,
		},
		{
			PackageSlug:       "sedan",
			TotalPriceInCents: 350,
		},
		{
			PackageSlug:       "van",
			TotalPriceInCents: 400,
		},
		{
			PackageSlug:       "luxury",
			TotalPriceInCents: 1000,
		},
	}
}

func (s *TripService) GetTripbyId(ctx context.Context, tripId string) (*domain.TripModel, error) {
	return s.Repository.GetTripbyId(ctx, tripId)
}
func (s *TripService) UpdateTrip(ctx context.Context, tripid string, status string, driver *pbd.Driver) error {
	return s.Repository.UpdateTrip(ctx, tripid, status, driver)
}
