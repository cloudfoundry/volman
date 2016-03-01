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

type LocalClient struct {
	DriversPath string
}

func NewLocalClient(driversPath string) *LocalClient {
	return &LocalClient{driversPath}
}

func (client *LocalClient) ListDrivers(logger lager.Logger) (volman.ListDriversResponse, error) {
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

func (client *LocalClient) listDrivers(logger lager.Logger) ([]voldriver.InfoResponse, error) {
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
		driver := voldriver.NewDriverClientCli(client.DriversPath, &system.SystemExec{}, driverInfoResponse.Name)
		InfoResponse, err := driver.Info(logger)
		if err != nil {
			return nil, fmt.Errorf(" Error occured in list drivers (%s)", err.Error())
		}
		drivers = append(drivers, InfoResponse)
	}
	return drivers, nil
}

func (client *LocalClient) Mount(logger lager.Logger, driverId string, volumeId string, config string) (volman.MountResponse, error) {
	logger = logger.Session("mount")
	logger.Info("start")
	defer logger.Info("end")

	drivers, err := client.listDrivers(logger)
	if err != nil {
		return volman.MountResponse{}, fmt.Errorf("Volman cannot find drivers ", err.Error())
	}
	for _, driverInfoResponse := range drivers {
		if driverInfoResponse.Name == driverId {
			driver := voldriver.NewDriverClientCli(client.DriversPath, &system.SystemExec{}, driverInfoResponse.Name)
			mountResponse, err := driver.Mount(logger, voldriver.MountRequest{VolumeId: volumeId, Config: config})
			if err != nil {
				return volman.MountResponse{}, err
			}
			return volman.MountResponse{mountResponse.Path}, nil
		}
	}
	return volman.MountResponse{}, fmt.Errorf("Driver 'InvalidDriver' not found in list of known drivers")

}
