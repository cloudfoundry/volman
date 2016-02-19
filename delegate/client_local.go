package delegate

import (
	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/system"
	"github.com/pivotal-golang/lager"
)

type LocalClient struct {
	UseDriverClient DriverClient
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

	return volman.ListDriversResponse{drivers}, nil
}

func (client *LocalClient) Mount(logger lager.Logger, driver volman.Driver, volumeId string, config string) (volman.MountPointResponse, error) {
	logger = logger.Session("mount")
	logger.Info("start")
	defer logger.Info("end")

	mountPath, err := client.UseDriverClient.Mount(logger, driver, volumeId, config)
	if err != nil {
		return volman.MountPointResponse{}, err
	}
	return volman.MountPointResponse{mountPath}, nil
}
