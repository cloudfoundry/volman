package delegate

import (
	"fmt"
	"path/filepath"

	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/delegate/driverclient"
	"github.com/cloudfoundry-incubator/volman/system"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter -o ../volmanfakes/fake_driver_client.go . DriverClient

type DriverClient interface {
	ListDrivers(logger lager.Logger) ([]volman.Driver, error)
	Mount(logger lager.Logger, driver volman.Driver, volumeId string, config string) (string, error)
}

type DriverClientCli struct {
	UseExec     system.Exec
	DriversPath string
}

func NewDriverClientCli(path string, useExec system.Exec) DriverClient {
	return &DriverClientCli{
		UseExec:     useExec,
		DriversPath: path,
	}
}

func (client *DriverClientCli) ListDrivers(logger lager.Logger) ([]volman.Driver, error) {
	driversBinaries, err := filepath.Glob(client.DriversPath + "/*")
	if err != nil { // untestable on linux, does glob work differently on windows???
		return nil, fmt.Errorf("Volman configured with an invalid driver path '%s', error occured list files (%s)", client.DriversPath, err.Error())
	}
	logger.Info(fmt.Sprintf("Found binaries: %#v", driversBinaries))

	listing := driverclient.Listing{driversBinaries, client.UseExec}
	return listing.List(logger, client.DriversPath)
}

func (client *DriverClientCli) Mount(logger lager.Logger, driver volman.Driver, volumeId string, config string) (string, error) {
	driverList, err := client.ListDrivers(logger)
	if err != nil {
		logger.Info(fmt.Sprintf("List Drivers fails: %s", err.Error()))
		return "", err
	}
	mounting := driverclient.Mounting{driver, driverList, client.UseExec}
	return mounting.Mount(logger, volumeId, config)
}
