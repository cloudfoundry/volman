package vollocal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/cloudfoundry-incubator/volman/voldriver/driverhttp"
	"github.com/pivotal-golang/lager"
)

type localClient struct {
	DriversPath string
	Factory     driverhttp.RemoteClientFactory
}

func NewLocalClient(driversPath string) *localClient {
	remoteClientFactory := driverhttp.NewRemoteClientFactory()
	return NewLocalClientWithRemoteClientFactory(driversPath, remoteClientFactory)
}

func NewLocalClientWithRemoteClientFactory(driversPath string, factory driverhttp.RemoteClientFactory) *localClient {
	return &localClient{driversPath, factory}
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

	driversBinaries, err := filepath.Glob(client.DriversPath + "/*.json")
	if err != nil { // untestable on linux, does glob work differently on windows???
		return nil, fmt.Errorf("Volman configured with an invalid driver path '%s', error occured list files (%s)", client.DriversPath, err.Error())
	}

	logger.Info("found-json-specs", lager.Data{"driversBinaries": driversBinaries})
	drivers := []voldriver.InfoResponse{}

	for _, driverExecutable := range driversBinaries {
		split := strings.Split(driverExecutable, "/")
		driverInfoResponse := voldriver.InfoResponse{Name: strings.TrimSuffix(split[len(split)-1], ".json")}

		drivers = append(drivers, driverInfoResponse)
	}
	return drivers, nil
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

	var response voldriver.MountResponse
	err = client.callDriver(logger, driverId, func(driver voldriver.Driver) error {
		mountRequest := voldriver.MountRequest{Name: volumeId}
		logger.Info("calling-driver-with-mount request", lager.Data{"driverId": driverId, "mountRequest": mountRequest})
		response = driver.Mount(logger, mountRequest)

		logger.Info("response-from-driver", lager.Data{"response": response})

		if response.Err == "" {
			return nil
		}
		return errors.New(response.Err)
	})

	return volman.MountResponse{response.Mountpoint}, err
}

func (client *localClient) Unmount(logger lager.Logger, driverId string, volumeName string) error {
	logger = logger.Session("unmount")
	logger.Info("start")
	logger.Info("unmounting-volume", lager.Data{"volumeName": volumeName})
	defer logger.Info("end")

	err := client.callDriver(logger, driverId, func(driver voldriver.Driver) error {
		response := driver.Unmount(logger, voldriver.UnmountRequest{Name: volumeName})
		if response.Err == "" {
			return nil
		}
		return errors.New(response.Err)
	})

	return err
}

func (client *localClient) create(logger lager.Logger, driverId string, volumeName string, opts map[string]interface{}) error {
	logger = logger.Session("create")
	logger.Info("start")
	defer logger.Info("end")

	logger.Info("creating-volume", lager.Data{"volumeName": volumeName, "driverId": driverId, "opts": opts})
	err := client.callDriver(logger, driverId, func(driver voldriver.Driver) error {
		response := driver.Create(logger, voldriver.CreateRequest{Name: volumeName, Opts: opts})

		if response.Err == "" {
			return nil
		}
		return errors.New(response.Err)
	})

	return err
}

type driverCallback func(driver voldriver.Driver) error

func (client *localClient) callDriver(logger lager.Logger, driverId string, callback driverCallback) error {
	drivers, err := client.listDrivers(logger)
	if err != nil {
		return fmt.Errorf("Volman cannot find any drivers", err.Error())
	}
	var driver voldriver.Driver
	for _, driverInfoResponse := range drivers {
		if driverInfoResponse.Name == driverId {
			// extract url from json file

			var driverJsonSpec voldriver.DriverSpec
			configFile, err := os.Open(client.DriversPath + "/" + driverInfoResponse.Name + ".json")
			if err != nil {
				fmt.Errorf("opening config file", err.Error())
			}

			jsonParser := json.NewDecoder(configFile)
			if err = jsonParser.Decode(&driverJsonSpec); err != nil {
				logger.Error("parsing-config-file-error", err)
				return err
			}

			logger.Info("invoking-driver", lager.Data{"address": driverJsonSpec.Address})
			driver, _ = client.Factory.NewRemoteClient(driverJsonSpec.Address)
			err = callback(driver)
			if err != nil {
				logger.Error(fmt.Sprintf("error-calling-driver%s-error-%#v", driverId), err)
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("Driver '%s' not found in list of known drivers", driverId)
}
