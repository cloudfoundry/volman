package vollocal

import (
	"errors"

	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/pivotal-golang/lager"
)

type localClient struct {
	driverFactory DriverFactory
}

func NewLocalClient(driversPath string) *localClient {
	return NewLocalClientWithDriverFactory(NewDriverFactory(driversPath))
}

func NewLocalClientWithDriverFactory(driverFactory DriverFactory) *localClient {
	return &localClient{driverFactory}
}

func (client *localClient) ListDrivers(logger lager.Logger) (volman.ListDriversResponse, error) {
	logger = logger.Session("list-drivers")
	logger.Info("start")
	defer logger.Info("end")

	drivers, err := client.driverFactory.Discover(logger)
	if err != nil {
		return volman.ListDriversResponse{}, err
	}
	logger.Info("listing-drivers")
	var infoResponses []voldriver.InfoResponse
	for driverName, driverFileName := range drivers {
		logger.Info("driver-name", lager.Data{"drivername": driverName, "driverfilename": driverFileName})
		infoResponses = append(infoResponses, voldriver.InfoResponse{driverName, driverFileName})
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

	driver, err := client.driverFactory.Driver(logger, driverId)
	if err != nil {
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

	driver, err := client.driverFactory.Driver(logger, driverId)
	if err != nil {
		logger.Error("mount-driver-lookup-error", err)
		return err
	}

	if response := driver.Unmount(logger, voldriver.UnmountRequest{Name: volumeName}); response.Err != "" {
		logger.Error("unmount-failed", err)
		return errors.New(response.Err)
	}

	if response := driver.Remove(logger, voldriver.RemoveRequest{Name: volumeName}); response.Err != "" {
		logger.Error("remove-failed", err)
		return errors.New(response.Err)
	}

	return nil
}

func (client *localClient) create(logger lager.Logger, driverId string, volumeName string, opts map[string]interface{}) error {
	logger = logger.Session("create")
	logger.Info("start")
	defer logger.Info("end")

	driver, err := client.driverFactory.Driver(logger, driverId)
	if err != nil {
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
