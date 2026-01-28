package domain

import (
	tripTypes "ride-sharing/services/trip-service/pkg/types"
	pb "ride-sharing/shared/proto/trip"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RideFareModel struct {
	ID                primitive.ObjectID
	UserID            string
	PackageSlug       string
	TotalPriceInCents float64
	ExpiresAt         time.Time
	Route             *tripTypes.OsmrApiResponse
}

func (f *RideFareModel) ToProto() *pb.RideFare {
	return &pb.RideFare{
		Id:                f.ID.Hex(),
		Userid:            f.UserID,
		PackageSlug:       f.PackageSlug,
		TotalPriceinCents: f.TotalPriceInCents,
	}
}

func ToRideFaresProto(fares []*RideFareModel) []*pb.RideFare {
	res := make([]*pb.RideFare, len(fares))

	for _, f := range fares {
		res = append(res, f.ToProto())
	}

	return res
}
