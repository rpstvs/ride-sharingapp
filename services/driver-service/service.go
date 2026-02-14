package main

import (
	"log"
	math "math/rand/v2"
	pb "ride-sharing/shared/proto/driver"
	"ride-sharing/shared/util"
	"sync"

	"github.com/mmcloughlin/geohash"
)

type Service struct {
	Drivers []*DriverInMap
	mu      sync.RWMutex
}

type DriverInMap struct {
	Driver *pb.Driver
}

func NewService() *Service {
	return &Service{
		Drivers: make([]*DriverInMap, 0),
	}
}

//implement register and unregister.

func (s *Service) RegisterDriver(driverId string, packageSlug string) (*pb.Driver, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	randomIndex := math.IntN(len(PredefinedRoutes))
	randomRoute := PredefinedRoutes[randomIndex]
	randomAvatar := util.GetRandomAvatar(randomIndex)
	randomPlate := GenerateRandomPlate()
	geohash := geohash.Encode(randomRoute[0][0], randomRoute[0][1])

	driver2 := &pb.Driver{
		GeoHashPos:     geohash,
		Location:       &pb.Location{Latitude: randomRoute[0][0], Longitude: randomRoute[0][1]},
		Name:           "Lewis Hamilton",
		Id:             driverId,
		PackageSlug:    packageSlug,
		ProfilePicture: randomAvatar,
		CarPlate:       randomPlate,
	}

	s.Drivers = append(s.Drivers, &DriverInMap{Driver: driver2})

	return driver2, nil

}

func (s *Service) UnregisterDriver(driverId string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, driver := range s.Drivers {
		if driverId == driver.Driver.Id {
			s.Drivers[i] = s.Drivers[len(s.Drivers)-1]
			s.Drivers = s.Drivers[:len(s.Drivers)-1]
		}
	}
}

func (s *Service) FindAvailableDrivers(packageSlug string) []string {
	drivers := []string{}

	for _, driver := range s.Drivers {
		log.Println(driver)
		if driver.Driver.PackageSlug == packageSlug {
			drivers = append(drivers, driver.Driver.Id)
		}
	}

	if len(drivers) == 0 {
		return []string{}
	}
	log.Printf("drivers in map %v", drivers)
	return drivers
}
