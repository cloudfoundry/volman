package vollocal

import (
	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/system"
	"github.com/pivotal-golang/lager"
)

type LocalClient struct {
	UseDriverClient volman.DriverPlugin
}

func NewLocalClient(driversPath string) *LocalClient {
	driverClientCli := NewDriverClientCli(driversPath, &system.SystemExec{})
	return &LocalClient{driverClientCli}
}

func (client *LocalClient) ListDrivers(logger lager.Logger) (volman.ListDriversResponse, error) {
	logger = logger.Session("list-drivers")
	logger.Info("start")
	defer logger.Info("end")

	drivers, err := client.UseDriverClient.ListDrivers(logger)
	if err != nil {
		return volman.ListDriversResponse{}, err
	}

	var driverInfos []volman.DriverInfo
	for _, driver := range drivers {
		driverInfos = append(driverInfos, volman.DriverInfo{driver.Name})
	}

	return volman.ListDriversResponse{driverInfos}, nil
}

func (client *LocalClient) Mount(logger lager.Logger, driverId string, volumeId string, config string) (volman.MountPointResponse, error) {
	logger = logger.Session("mount")
	logger.Info("start")
	defer logger.Info("end")

	driver := volman.Driver{Name: driverId}

	mountPath, err := client.UseDriverClient.Mount(logger, driver, volumeId, config)
	if err != nil {
		return volman.MountPointResponse{}, err
	}
	return volman.MountPointResponse{mountPath}, nil
}
