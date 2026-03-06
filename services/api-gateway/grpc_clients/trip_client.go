package grpc_clients

import (
	"os"
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/tracing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TripServiceClient struct {
	Client pb.TripServiceClient
	Conn   *grpc.ClientConn
}

func NewTripServiceClient() (*TripServiceClient, error) {
	tripServiceUrl := os.Getenv("TRIP_SERVICE_URL")
	if tripServiceUrl == "" {
		tripServiceUrl = "trip-service:9093"
	}

	dialOptions := append(tracing.DialOptionsWithTracing(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient(tripServiceUrl, dialOptions...)

	if err != nil {
		return nil, err
	}
	client := pb.NewTripServiceClient(conn)
	return &TripServiceClient{
		Conn:   conn,
		Client: client,
	}, nil

}

func (c *TripServiceClient) Close() {
	if c.Conn != nil {
		if err := c.Conn.Close(); err != nil {
			return
		}
	}
}
