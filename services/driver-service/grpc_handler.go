package main

import (
	"context"
	"log"
	pb "ride-sharing/shared/proto/driver"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type driverGrpcHandler struct {
	pb.UnimplementedDriverServiceServer
	service *Service
}

func NewgRPCHandler(s *grpc.Server, svc *Service) {
	handler := &driverGrpcHandler{
		service: svc,
	}
	pb.RegisterDriverServiceServer(s, handler)

}

func (h *driverGrpcHandler) RegisterDriver(ctx context.Context, req *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {

	driver, err := h.service.RegisterDriver(req.GetID(), req.GetPackageSlug())

	if err != nil {
		log.Printf("error registering driver %v", err)
		return nil, status.Errorf(codes.Internal, "error registering driver %v", err)
	}

	resp := &pb.RegisterDriverResponse{
		Driver: driver,
	}

	return resp, nil
}

func (h *driverGrpcHandler) UnregisterDriver(ctx context.Context, req *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {

	h.service.UnregisterDriver(req.GetID())

	return &pb.RegisterDriverResponse{
		Driver: &pb.Driver{
			Id: req.GetID(),
		},
	}, nil
}
