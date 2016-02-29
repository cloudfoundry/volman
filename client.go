package volman

import "github.com/pivotal-golang/lager"

type Manager interface {
	ListDrivers(logger lager.Logger) (ListDriversResponse, error)
	Mount(logger lager.Logger, driverId string, volumeId string, config string) (MountPointResponse, error)
}

//go:generate counterfeiter -o volmanfakes/fake_driver_client.go . DriverPlugin

type DriverPlugin interface {
	Info(logger lager.Logger) (DriverInfo, error)
	Mount(logger lager.Logger, volumeId string, config string) (string, error)
}
