package vollocal

import (
	"errors"
	"flag"
	"time"

	"github.com/tedsuo/ifrit"

	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"

	cf_lager "github.com/cloudfoundry-incubator/cf-lager"
)

type localClient struct {
	driverSyncer DriverSyncer
}

func NewLocalClient(logger lager.Logger, driversPath string) (*localClient, ifrit.Runner) {
	driverFactory := NewDriverFactory(driversPath)

	scanInterval := 30 * time.Second
	clock := clock.NewClock()

	return NewLocalClientWithDriverFactory(logger, driversPath, driverFactory, NewDriverSyncer(logger, driverFactory, scanInterval, clock))
}

func NewLocalClientWithDriverFactory(logger lager.Logger, driversPath string, driverFactory DriverFactory, driverSyncer DriverSyncer) (*localClient, ifrit.Runner) {
	cf_lager.AddFlags(flag.NewFlagSet("", flag.PanicOnError))
	flag.Parse()

	return &localClient{
		driverSyncer: driverSyncer,
	}, driverSyncer.Runner()
}

func (client *localClient) ListDrivers(logger lager.Logger) (volman.ListDriversResponse, error) {
	logger = logger.Session("list-drivers")
	logger.Debug("start")
	defer logger.Debug("end")

	var infoResponses []volman.InfoResponse
	drivers := client.driverSyncer.Drivers()

	for name, _ := range drivers {
		info := volman.InfoResponse{
			Name: name,
		}
		infoResponses = append(infoResponses, info)
	}
	logger.Debug("listing-drivers", lager.Data{"drivers": infoResponses})
	return volman.ListDriversResponse{infoResponses}, nil
}

func (client *localClient) Mount(logger lager.Logger, driverId string, volumeId string, config map[string]interface{}) (volman.MountResponse, error) {
	logger = logger.Session("mount")
	logger.Info("start")
	logger.Info("driver-mounting-volume", lager.Data{"driverId": driverId, "volumeId": volumeId})
	defer logger.Info("end")

	driver, found := client.driverSyncer.Driver(driverId)
	if !found {
		err := errors.New("Driver '" + driverId + "' not found in list of known drivers")
		logger.Error("mount-driver-lookup-error", err)
		return volman.MountResponse{}, err
	}

	err := client.activate(logger, driverId, driver)
	if err != nil {
		logger.Error("activate-failed", err)
		return volman.MountResponse{}, err
	}

	err = client.create(logger, driverId, volumeId, config)
	if err != nil {
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

	driver, found := client.driverSyncer.Driver(driverId)
	if !found {
		err := errors.New("Driver '" + driverId + "' not found in list of known drivers")
		logger.Error("mount-driver-lookup-error", err)
		return err
	}

	err := client.activate(logger, driverId, driver)
	if err != nil {
		logger.Error("activate-failed", err)
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

func (client *localClient) activate(logger lager.Logger, driverId string, driver voldriver.Driver) error {
	logger = logger.Session("activate")
	logger.Info("start")
	defer logger.Info("end")

	activated, err := client.driverSyncer.DriverRegistry().Activated(driverId)
	if err != nil {
		logger.Error("activate-driver-lookup-error", err)
		return err
	}

	if !activated {
		activateResponse := driver.Activate(logger)
		if activateResponse.Err != "" {
			logger.Error("driver-activate-error", err)
			return errors.New(activateResponse.Err)
		}

		err := client.driverSyncer.DriverRegistry().Activate(driverId)
		if err != nil {
			logger.Error("driver-registry-activate-error", err)
			return err
		}
		logger.Info("driver-activated", lager.Data{"driver": driverId})
	}

	return nil
}

func (client *localClient) create(logger lager.Logger, driverId string, volumeName string, opts map[string]interface{}) error {
	logger = logger.Session("create")
	logger.Info("start")
	defer logger.Info("end")
	driver, found := client.driverSyncer.Driver(driverId)
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
