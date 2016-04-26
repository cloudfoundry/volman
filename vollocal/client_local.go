package vollocal

import (
	"errors"
	"time"

	"github.com/tedsuo/ifrit"

	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
)

type localClient struct {
	driverRegistry DriverRegistry
	driverSyncer   DriverSyncer
}

func NewLocalClient(logger lager.Logger, driversPath string) (*localClient, ifrit.Runner) {
	driverFactory := NewDriverFactory(driversPath)

	scanInterval := 30 * time.Second
	clock := clock.NewClock()

	registry := NewDriverRegistry()
	syncer := NewDriverSyncer(logger, driverFactory, registry, scanInterval, clock)

	return NewLocalClientWithDriverFactory(logger, driversPath, driverFactory, syncer, registry)
}

func NewLocalClientWithDriverFactory(logger lager.Logger, driversPath string, driverFactory DriverFactory, driverSyncer DriverSyncer) (*localClient, ifrit.Runner) {
	return &localClient{
		driverSyncer:   driverSyncer,
		driverRegistry: driverRegistry,
	}, driverSyncer.Runner()
}

func (client *localClient) ListDrivers(logger lager.Logger) (volman.ListDriversResponse, error) {
	logger = logger.Session("list-drivers")
	logger.Info("start")
	defer logger.Info("end")

	var infoResponses []volman.InfoResponse
	drivers := client.driverRegistry.Drivers()

	for name, _ := range drivers {
		infoResponses = append(infoResponses, volman.InfoResponse{Name: name})
	}
	logger.Debug("listing-drivers", lager.Data{"drivers": infoResponses})
	return volman.ListDriversResponse{infoResponses}, nil
}

func (client *localClient) Mount(logger lager.Logger, driverId string, volumeId string, config map[string]interface{}) (volman.MountResponse, error) {
	logger = logger.Session("mount")
	logger.Info("start")
	defer logger.Info("end")
	logger.Debug("driver-mounting-volume", lager.Data{"driverId": driverId, "volumeId": volumeId})

	driver, found := client.driverRegistry.Driver(driverId)
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
	logger.Debug("calling-driver-with-mount-request", lager.Data{"driverId": driverId, "mountRequest": mountRequest})
	mountResponse := driver.Mount(logger, mountRequest)
	logger.Debug("response-from-driver", lager.Data{"response": mountResponse})
	if mountResponse.Err != "" {
		return volman.MountResponse{}, errors.New(mountResponse.Err)
	}

	return volman.MountResponse{mountResponse.Mountpoint}, nil
}

func (client *localClient) Unmount(logger lager.Logger, driverId string, volumeName string) error {
	logger = logger.Session("unmount")
	logger.Info("start")
	defer logger.Info("end")
	logger.Debug("unmounting-volume", lager.Data{"volumeName": volumeName})

	driver, found := client.driverRegistry.Driver(driverId)
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

	activated, err := client.driverRegistry.Activated(driverId)
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

		driverImplementsErr := errors.New(fmt.Sprintf("driver-implements: %#v", activateResponse.Implements))
		if len(activateResponse.Implements) == 0 {
			logger.Error("driver-incorrect", driverImplementsErr)
			return driverImplementsErr
		}

		if !client.driverImplements("VolumeDriver", activateResponse.Implements) {
			logger.Error("driver-incorrect", driverImplementsErr)
			return driverImplementsErr
		} else {
			err := client.driverRegistry.Activate(driverId)
			if err != nil {
				logger.Error("driver-registry-activate-error", err)
				return err
			}
			logger.Debug("driver-activated", lager.Data{"driver": driverId})
		}
	}

	return nil
}

func (client *localClient) create(logger lager.Logger, driverId string, volumeName string, opts map[string]interface{}) error {
	logger = logger.Session("create")
	logger.Info("start")
	defer logger.Info("end")
	driver, found := client.driverRegistry.Driver(driverId)
	if !found {
		err := errors.New("Driver '" + driverId + "' not found in list of known drivers")
		logger.Error("mount-driver-lookup-error", err)
		return err
	}

	logger.Debug("creating-volume", lager.Data{"volumeName": volumeName, "driverId": driverId, "opts": opts})
	response := driver.Create(logger, voldriver.CreateRequest{Name: volumeName, Opts: opts})
	if response.Err != "" {
		return errors.New(response.Err)
	}
	return nil
}

func (client *localClient) driverImplements(protocol string, activateResponseProtocols []string) bool {
	for _, nextProtocol := range activateResponseProtocols {
		if protocol == nextProtocol {
			return true
		}
	}
	return false
}
