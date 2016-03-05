package vollocal

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/system"
	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/pivotal-golang/lager"
)

type localClient struct {
	DriversPath string
	UseExec     system.Exec
}

func NewLocalClient(driversPath string) *localClient {
	return NewLocalClientWithExec(driversPath, &system.SystemExec{})
}

func NewLocalClientWithExec(driversPath string, exec system.Exec) *localClient {
	return &localClient{driversPath, exec}
}

func (client *localClient) ListDrivers(logger lager.Logger) (volman.ListDriversResponse, error) {
	logger = logger.Session("list-drivers")
	logger.Info("start")
	defer logger.Info("end")

	drivers, err := client.listDrivers(logger)

	if err != nil {
		return volman.ListDriversResponse{}, err
	}

	var InfoResponses []voldriver.InfoResponse
	for _, driver := range drivers {
		InfoResponses = append(InfoResponses, voldriver.InfoResponse{driver.Name, client.DriversPath})
	}

	return volman.ListDriversResponse{InfoResponses}, nil
}

func (client *localClient) listDrivers(logger lager.Logger) ([]voldriver.InfoResponse, error) {
	logger = logger.Session("list-drivers")
	logger.Info("start")
	defer logger.Info("end")
	driversBinaries, err := filepath.Glob(client.DriversPath + "/*")
	if err != nil { // untestable on linux, does glob work differently on windows???
		return nil, fmt.Errorf("Volman configured with an invalid driver path '%s', error occured list files (%s)", client.DriversPath, err.Error())
	}
	logger.Info(fmt.Sprintf("Found binaries: %#v", driversBinaries))
	drivers := []voldriver.InfoResponse{}

	for _, driverExecutable := range driversBinaries {
		split := strings.Split(driverExecutable, "/")
		driverInfoResponse := voldriver.InfoResponse{Name: split[len(split)-1]}
		driver := voldriver.NewDriverClientCli(client.DriversPath, client.UseExec, driverInfoResponse.Name)
		InfoResponse, err := driver.Info(logger)
		if err != nil {
			msg := fmt.Sprintf(" Error occured in list drivers for executable %s (%s)", driverExecutable, err.Error())
			logger.Error(msg, err)
			return nil, fmt.Errorf(msg)
		}
		drivers = append(drivers, InfoResponse)
	}
	return drivers, nil
}

func (client *localClient) Mount(logger lager.Logger, driverId string, volumeId string, config string) (volman.MountResponse, error) {
	logger = logger.Session("mount")
	logger.Info("start")
	defer logger.Info("end")

	var response voldriver.MountResponse
	err := client.driverCall(logger, driverId, func(driver voldriver.Driver) error {
		var err error
		response, err = driver.Mount(logger, voldriver.MountRequest{VolumeId: volumeId, Config: config})
		return err
	})
	return volman.MountResponse{response.Path}, err
}

func (client *localClient) Unmount(logger lager.Logger, driverId string, volumeId string) error {
	logger = logger.Session("unmount")
	logger.Info("start")
	defer logger.Info("end")

	err := client.driverCall(logger, driverId, func(driver voldriver.Driver) error {
		return driver.Unmount(logger, voldriver.UnmountRequest{VolumeId: volumeId})
	})
	return err
}

type driverCallback func(driver voldriver.Driver) error

func (client *localClient) driverCall(logger lager.Logger, driverId string, callback driverCallback) error {
	drivers, err := client.listDrivers(logger)
	if err != nil {
		return fmt.Errorf("Volman cannot find any drivers", err.Error())
	}
	var driver voldriver.Driver
	for _, driverInfoResponse := range drivers {
		if driverInfoResponse.Name == driverId {
			driver = voldriver.NewDriverClientCli(client.DriversPath, client.UseExec, driverInfoResponse.Name)
			err := callback(driver)
			if err != nil {
				logger.Error(fmt.Sprintf("Error calling driver %s", driverId), err)
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("Driver '%s' not found in list of known drivers", driverId)
}
