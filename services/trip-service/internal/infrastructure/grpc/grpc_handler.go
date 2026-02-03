package grpc

import (
	"context"
	"log"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/internal/infrastructure/events"
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHandler struct {
	pb.UnimplementedTripServiceServer
	service   domain.TripService
	publisher *events.TripEventPublisher
}

func NewgRPCHandler(server *grpc.Server, service domain.TripService, publisher *events.TripEventPublisher) *gRPCHandler {
	handler := &gRPCHandler{
		service:   service,
		publisher: publisher,
	}

	pb.RegisterTripServiceServer(server, handler)

	return handler
}

func (h *gRPCHandler) PreviewTrip(ctx context.Context, r *pb.PreviewTripRequest) (*pb.PreviewTripResponse, error) {
	pickup := types.Coordinate{
		Latitude:  r.GetStartLocation().Latitude,
		Longitude: r.GetStartLocation().Longitude,
	}

	dest := types.Coordinate{
		Latitude:  r.GetEndLocation().Latitude,
		Longitude: r.GetEndLocation().Longitude,
	}

	response, err := h.service.GetRoute(ctx, pickup, dest)

	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to get route %v", err)
	}

	estimatedFares := h.service.EstimatePackagesPricesWithRoute(response)

	fares, err := h.service.GenerateTripFares(ctx, estimatedFares, r.UserID, response)

	if err != nil {
		return &pb.PreviewTripResponse{}, status.Errorf(codes.Internal, "failed to generate ride fares %v", err)
	}

	return &pb.PreviewTripResponse{
		Route:     response.ToProto(),
		RideFares: domain.ToRideFaresProto(fares),
	}, nil
}

func (h *gRPCHandler) StartTrip(ctx context.Context, r *pb.CreateTripRequest) (*pb.CreateTripResponse, error) {
	userid := r.GetUserID()
	fareid := r.GetRideFareId()
	fare, err := h.service.GetAndValidateFare(ctx, fareid, userid)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "error validate fare: %v", err)
	}

	trip, err := h.service.CreateTrip(ctx, fare)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "error creating trip: %v", err)
	}

	if err := h.publisher.PublishTripCreated(ctx); err != nil {
		return nil, err
	}

	return &pb.CreateTripResponse{
		TripId: trip.ID.Hex(),
	}, nil

}
