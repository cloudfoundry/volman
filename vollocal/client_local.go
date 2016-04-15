package vollocal

import (
	"errors"
	"flag"
	"time"

	cf_lager "github.com/cloudfoundry-incubator/cf-lager"
	"github.com/tedsuo/ifrit"

	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
)

type localClient struct {
	Registry *DriverSyncer
}

func NewLocalClient(logger lager.Logger, driversPath string) (*localClient, ifrit.Runner) {
	driverFactory := NewDriverFactory(driversPath)
	return NewLocalClientWithDriverFactory(logger, driversPath, driverFactory)
}

func NewLocalClientWithDriverFactory(logger lager.Logger, driversPath string, driverFactory DriverFactory) (*localClient, ifrit.Runner) {
	cf_lager.AddFlags(flag.NewFlagSet("", flag.PanicOnError))
	flag.Parse()

	scanInterval := 30 * time.Second
	clock := clock.NewClock()
	registry := NewDriverSyncer(logger, driverFactory, scanInterval, clock)

	return &localClient{registry}, registry
}

func (client *localClient) ListDrivers(logger lager.Logger) (volman.ListDriversResponse, error) {
	logger = logger.Session("list-drivers")
	logger.Debug("start")
	defer logger.Debug("end")

	logger.Debug("listing-drivers")
	var infoResponses []voldriver.InfoResponse
	drivers := client.Registry.Drivers()

	for _, driver := range drivers {

		info, err := driver.Info(logger)
		if err != nil {
			return volman.ListDriversResponse{}, err
		}
		infoResponses = append(infoResponses, info)
	}
	return volman.ListDriversResponse{infoResponses}, nil
}

func (client *localClient) Mount(logger lager.Logger, driverId string, volumeId string, config map[string]interface{}) (volman.MountResponse, error) {
	logger = logger.Session("mount")
	logger.Info("start")
	logger.Info("driver-mounting-volume", lager.Data{"driverId": driverId, "volumeId": volumeId})
	defer logger.Info("end")

	err := client.create(logger, driverId, volumeId, config)
	if err != nil {
		return volman.MountResponse{}, err
	}

	driver, found := client.Registry.Driver(driverId)
	if !found {
		err := errors.New("Driver '" + driverId + "' not found in list of known drivers")
		logger.Error("mount-driver-lookup-error", err)
		return volman.MountResponse{}, err
	}
	mountRequest := voldriver.MountRequest{Name: volumeId}
	logger.Info("calling-driver-with-mount-request", lager.Data{"driverId": driverId, "mountRequest": mountRequest})
	mountResponse := driver.Mount(logger, mountRequest)
	logger.Info("response-from-driver", lager.Data{"response": mountResponse})
	if mountResponse.Err != "" {
		return volman.MountResponse{}, errors.New(mountResponse.Err)
	}

	return volman.MountResponse{mountResponse.Mountpoint}, nil
}

func (client *localClient) Unmount(logger lager.Logger, driverId string, volumeName string) error {
	logger = logger.Session("unmount")
	logger.Info("start")
	logger.Info("unmounting-volume", lager.Data{"volumeName": volumeName})
	defer logger.Info("end")

	driver, found := client.Registry.Driver(driverId)
	if !found {
		err := errors.New("Driver '" + driverId + "' not found in list of known drivers")
		logger.Error("mount-driver-lookup-error", err)
		return err
	}

	if response := driver.Unmount(logger, voldriver.UnmountRequest{Name: volumeName}); response.Err != "" {
		err := errors.New(response.Err)
		logger.Error("unmount-failed", err)
		return err
	}

	if response := driver.Remove(logger, voldriver.RemoveRequest{Name: volumeName}); response.Err != "" {
		err := errors.New(response.Err)
		logger.Error("remove-failed", err)
		return err
	}

	return nil
}

func (client *localClient) create(logger lager.Logger, driverId string, volumeName string, opts map[string]interface{}) error {
	logger = logger.Session("create")
	logger.Info("start")
	defer logger.Info("end")
	driver, found := client.Registry.Driver(driverId)
	if !found {
		err := errors.New("Driver '" + driverId + "' not found in list of known drivers")
		logger.Error("mount-driver-lookup-error", err)
		return err
	}

	logger.Info("creating-volume", lager.Data{"volumeName": volumeName, "driverId": driverId, "opts": opts})
	response := driver.Create(logger, voldriver.CreateRequest{Name: volumeName, Opts: opts})
	if response.Err != "" {
		return errors.New(response.Err)
	}
	return nil
}
