package volman

import "github.com/pivotal-golang/lager"

//go:generate counterfeiter -o volmanfakes/fake_manager_client.go . Manager

type Manager interface {
	ListDrivers(logger lager.Logger) (ListDriversResponse, error)
	Mount(logger lager.Logger, driverId string, volumeId string, config map[string]interface{}) (MountResponse, error)
	Unmount(logger lager.Logger, driverId string, volumeId string) error
}
