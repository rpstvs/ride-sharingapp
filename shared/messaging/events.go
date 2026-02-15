package messaging

import (
	pbd "ride-sharing/shared/proto/driver"
	pb "ride-sharing/shared/proto/trip"
)

const (
	FindAvailableDriversQueue       = "find_available_drivers"
	DriverCmdTripRequestQueue       = "driver_trip_request"
	DriverTripResponseQueue         = "driver_trip_response"
	NotifyDriverNoDriversFoundQueue = "notify_driver_no_drivers_found"
	NotifyDriverAssignedQueue       = "notify_driver_assign"
	PaymentTripResponseQueue        = "payment_trip_response"
	NotifySessionCreatedQueue       = "notify_payment_session_created"
	NotifyPaymentSuccessQueue       = "notify_payment_success"
)

type TripEvent struct {
	Trip *pb.Trip `json:"trip"`
}

type DriveTripResponseData struct {
	Driver  *pbd.Driver `json:"driver"`
	TripID  string      `json:"tripID"`
	RiderID string      `json:"riderID"`
}

type PaymentEventSessionCreatedData struct {
	TripID    string  `json:"tripID"`
	SessionID string  `json:"sessionID"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
}

type PaymentTripResponseData struct {
	TripID   string  `json:"tripID"`
	UserID   string  `json:"userID"`
	DriverID string  `json:"driverID"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type PaymentStatusUpdateData struct {
	TripID   string `json:"tripID"`
	UserID   string `json:"userID"`
	DriverID string `json:"driverID"`
}
