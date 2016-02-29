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

	var driverInfos []volman.DriverInfo
	for _, driver := range drivers {
		driverInfos = append(driverInfos, volman.DriverInfo{driver.Name, client.DriversPath})
	}

	return volman.ListDriversResponse{driverInfos}, nil
}

func (client *LocalClient) listDrivers(logger lager.Logger) ([]volman.DriverInfo, error) {
	logger.Info("start")
	defer logger.Info("end")
	driversBinaries, err := filepath.Glob(client.DriversPath + "/*")
	if err != nil { // untestable on linux, does glob work differently on windows???
		return nil, fmt.Errorf("Volman configured with an invalid driver path '%s', error occured list files (%s)", client.DriversPath, err.Error())
	}
	logger.Info(fmt.Sprintf("Found binaries: %#v", driversBinaries))
	drivers := []volman.DriverInfo{}

	for _, driverExecutable := range driversBinaries {
		split := strings.Split(driverExecutable, "/")
		driver := volman.DriverInfo{Name: split[len(split)-1]}
		driverPlugin := voldriver.NewDriverClientCli(client.DriversPath, &system.SystemExec{}, driver.Name)
		driverInfo, err := driverPlugin.Info(logger)
		if err != nil {
			return nil, fmt.Errorf(" Error occured in list drivers (%s)", err.Error())
		}
		drivers = append(drivers, driverInfo)
	}
	return drivers, nil
}

func (client *LocalClient) Mount(logger lager.Logger, driverId string, volumeId string, config string) (volman.MountPointResponse, error) {
	logger = logger.Session("mount")
	logger.Info("start")
	defer logger.Info("end")

	drivers, err := client.listDrivers(logger)
	if err != nil {
		return volman.MountPointResponse{}, fmt.Errorf("Volman cannot find drivers ", err.Error())
	}
	for _, driver := range drivers {
		if driver.Name == driverId {
			driverPlugin := voldriver.NewDriverClientCli(client.DriversPath, &system.SystemExec{}, driver.Name)
			mountPath, err := driverPlugin.Mount(logger, volumeId, config)
			if err != nil {
				return volman.MountPointResponse{}, err
			}
			return volman.MountPointResponse{mountPath}, nil
		}
	}
	return volman.MountPointResponse{}, fmt.Errorf("Driver 'InvalidDriver' not found in list of known drivers")

}
