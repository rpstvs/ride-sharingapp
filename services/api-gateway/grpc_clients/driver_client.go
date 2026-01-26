package grpc_clients

import (
	"os"
	pb "ride-sharing/shared/proto/driver"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type DriverServiceClient struct {
	Client pb.DriverServiceClient
	Conn   *grpc.ClientConn
}

func NewDriverServiceClient() (*DriverServiceClient, error) {
	url := os.Getenv("DRIVER_SERVICE_URL")

	if url == "" {
		url = "driver-service:9094"
	}

	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}
	client := pb.NewDriverServiceClient(conn)

	return &DriverServiceClient{
		Client: client,
		Conn:   conn,
	}, nil
}

func (c *DriverServiceClient) Close() {
	if c.Conn != nil {
		if err := c.Conn.Close(); err != nil {
			return
		}
	}
}
