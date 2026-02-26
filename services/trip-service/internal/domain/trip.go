package domain

import (
	"context"
	tripTypes "ride-sharing/services/trip-service/pkg/types"
	pbd "ride-sharing/shared/proto/driver"
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TripModel struct {
	ID       primitive.ObjectID
	UserID   string
	Status   string
	RideFare *RideFareModel
	Driver   *pb.TripDriver
}

func (t *TripModel) ToProto() *pb.Trip {
	return &pb.Trip{
		Id:           t.ID.Hex(),
		UserID:       t.UserID,
		SelectedFare: t.RideFare.ToProto(),
		Status:       t.Status,
		Driver:       t.Driver,
		Route:        t.RideFare.Route.ToProto(),
	}
}

type TripRepository interface {
	CreateTrip(ctx context.Context, trip *TripModel) (*TripModel, error)
	SaveRideFare(ctx context.Context, f *RideFareModel) error

	GetRideFareByID(ctx context.Context, id string) (*RideFareModel, error)
	GetTripbyId(ctx context.Context, tripId string) (*TripModel, error)
	UpdateTrip(ctx context.Context, tripid string, status string, driver *pbd.Driver) error
}

type TripService interface {
	CreateTrip(ctx context.Context, fare *RideFareModel) (*TripModel, error)
	GetRoute(ctx context.Context, pickup, dest types.Coordinate) (*tripTypes.OsmrApiResponse, error)
	EstimatePackagesPricesWithRoute(route *tripTypes.OsmrApiResponse) []*RideFareModel
	GenerateTripFares(ctx context.Context, fares []*RideFareModel, userid string, route *tripTypes.OsmrApiResponse) ([]*RideFareModel, error)

	GetAndValidateFare(ctx context.Context, fareId, Userid string) (*RideFareModel, error)
	GetTripbyId(ctx context.Context, tripId string) (*TripModel, error)
	UpdateTrip(ctx context.Context, tripid string, status string, driver *pbd.Driver) error
}
