package vollocal

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
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

	drivers, err := client.discoverDrivers(logger)
	if err != nil {
		return volman.ListDriversResponse{}, err
	}
	logger.Info("listing-drivers")
	var infoResponses []voldriver.InfoResponse
	for driverName, driverFileName := range drivers {
		logger.Info("driver-name", lager.Data{"drivername": driverName, "driverfilename": driverFileName})
		infoResponses = append(infoResponses, voldriver.InfoResponse{driverName, path.Join(client.DriversPath, driverFileName)})
	}

	return volman.ListDriversResponse{infoResponses}, nil
}

func (client *localClient) discoverDrivers(logger lager.Logger) (map[string]string, error) {
	logger = logger.Session("discover-drivers")
	logger.Info("start")
	defer logger.Info("end")
	//precedence order: sock -> spec -> json
	spec_types := [3]string{"sock", "spec", "json"}
	endpoints := make(map[string]string)
	for _, spec_type := range spec_types {
		matchingDriverSpecs, err := client.getMatchingDriverSpecs(logger, spec_type)
		if err != nil { // untestable on linux, does glob work differently on windows???
			return nil, fmt.Errorf("Volman configured with an invalid driver path '%s', error occured list files (%s)", client.DriversPath, err.Error())
		}
		logger.Info("driver-specs", lager.Data{"drivers": matchingDriverSpecs})
		endpoints = insertIfNotFound(logger, endpoints, matchingDriverSpecs)
	}
	logger.Info("found-specs", lager.Data{"endpoints": endpoints})
	return endpoints, nil
}

func insertIfNotFound(logger lager.Logger, endpoints map[string]string, specs []string) map[string]string {
	for _, spec := range specs {
		split := strings.Split(spec, "/")
		specFileName := split[len(split)-1]
		specName := strings.Split(specFileName, ".")[0]
		logger.Info("insert-unique-specs", lager.Data{"specname": specName, "specFileName": specFileName})
		_, ok := endpoints[specName]
		if ok == false {
			endpoints[specName] = specFileName
		}
	}
	logger.Info("insert-if-unique", lager.Data{"endpoints": endpoints})
	return endpoints
}
func (client *localClient) getMatchingDriverSpecs(logger lager.Logger, pattern string) ([]string, error) {
	matchingDriverSpecs, err := filepath.Glob(client.DriversPath + "/*." + pattern)
	if err != nil { // untestable on linux, does glob work differently on windows???
		return nil, fmt.Errorf("Volman configured with an invalid driver path '%s', error occured list files (%s)", client.DriversPath, err.Error())
	}
	logger.Info("binaries", lager.Data{"binaries": matchingDriverSpecs})
	return matchingDriverSpecs, nil

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
	drivers, err := client.discoverDrivers(logger)
	if err != nil {
		return fmt.Errorf("Volman cannot find any drivers", err.Error())
	}
	var driver voldriver.Driver
	for driverName, driverFileName := range drivers {
		if driverName == driverId {
			var address string
			if strings.Contains(driverFileName, ".") {
				extension := strings.Split(driverFileName, ".")[1]
				switch extension {
				case "sock":
					address = path.Join("unix://", client.DriversPath, driverFileName)
				case "spec":
					configFile, err := os.Open(path.Join(client.DriversPath, driverFileName))
					if err != nil {
						fmt.Errorf("opening config file", err.Error())
					}
					reader := bufio.NewReader(configFile)
					addressBytes, _, _ := reader.ReadLine()
					address = string(addressBytes)
				case "json":
					// extract url from json file
					var driverJsonSpec voldriver.DriverSpec
					configFile, err := os.Open(path.Join(client.DriversPath, driverFileName))
					if err != nil {
						fmt.Errorf("opening config file", err.Error())
					}

					jsonParser := json.NewDecoder(configFile)
					if err = jsonParser.Decode(&driverJsonSpec); err != nil {
						logger.Error("parsing-config-file-error", err)
						return err

					}
					address = driverJsonSpec.Address
				}

				logger.Info("invoking-driver", lager.Data{"address": address})
				driver, _ = client.Factory.NewRemoteClient(address)
				err = callback(driver)
				if err != nil {
					logger.Error(fmt.Sprintf("error-calling-driver%s-error-%#v", driverId), err)
					return err
				}
				return nil
			}
		}
	}
	return fmt.Errorf("Driver '%s' not found in list of known drivers", driverId)
}
