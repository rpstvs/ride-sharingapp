package messaging

import pb "ride-sharing/shared/proto/trip"

const (
	FindAvailableDriversQueue = "find_available_drivers"
	DriverCmdTripRequestQueue = "driver_trip_request"
)

type TripEvent struct {
	Trip *pb.Trip `json:"trip"`
}
