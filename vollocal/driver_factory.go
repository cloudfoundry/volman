package vollocal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/cloudfoundry-incubator/volman/voldriver/driverhttp"
	"github.com/pivotal-golang/lager"
)

//go:generate counterfeiter -o ../volmanfakes/fake_driver_factory.go . DriverFactory

type DriverFactory interface {
	Discover(logger lager.Logger) (map[string]string, error)
	Driver(logger lager.Logger, driverId string) (voldriver.Driver, error)
}

type realDriverFactory struct {
	DriversPath string
	Factory     driverhttp.RemoteClientFactory
}

func NewDriverFactory(driversPath string) DriverFactory {
	remoteClientFactory := driverhttp.NewRemoteClientFactory()
	return NewDriverFactoryWithRemoteClientFactory(driversPath, remoteClientFactory)
}

func NewDriverFactoryWithRemoteClientFactory(driversPath string, remoteClientFactory driverhttp.RemoteClientFactory) DriverFactory {
	return &realDriverFactory{driversPath, remoteClientFactory}
}

func (r *realDriverFactory) Discover(logger lager.Logger) (map[string]string, error) {
	logger = logger.Session("discover")
	logger.Info("start")
	defer logger.Info("end")
	//precedence order: sock -> spec -> json
	spec_types := [3]string{"sock", "spec", "json"}
	endpoints := make(map[string]string)
	for _, spec_type := range spec_types {
		matchingDriverSpecs, err := r.getMatchingDriverSpecs(logger, spec_type)
		if err != nil { // untestable on linux, does glob work differently on windows???
			return nil, fmt.Errorf("Volman configured with an invalid driver path '%s', error occured list files (%s)", r.DriversPath, err.Error())
		}
		logger.Info("driver-specs", lager.Data{"drivers": matchingDriverSpecs})
		endpoints = r.insertIfNotFound(logger, endpoints, matchingDriverSpecs)
	}
	logger.Info("found-specs", lager.Data{"endpoints": endpoints})
	return endpoints, nil
}

func (*realDriverFactory) insertIfNotFound(logger lager.Logger, endpoints map[string]string, specs []string) map[string]string {
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

func (r *realDriverFactory) Driver(logger lager.Logger, driverId string) (voldriver.Driver, error) {
	drivers, err := r.Discover(logger)
	if err != nil {
		return nil, fmt.Errorf("Volman cannot find any drivers", err.Error())
	}
	var driver voldriver.Driver
	for driverName, driverFileName := range drivers {
		if driverName == driverId {
			var address string
			if strings.Contains(driverFileName, ".") {
				extension := strings.Split(driverFileName, ".")[1]
				switch extension {
				case "sock":
					address = path.Join("unix://", r.DriversPath, driverFileName)
				case "spec":
					configFile, err := os.Open(path.Join(r.DriversPath, driverFileName))
					if err != nil {
						logger.Error(fmt.Sprintf("error-opening-config-%s", driverFileName), err)
						return nil, err
					}
					reader := bufio.NewReader(configFile)
					addressBytes, _, err := reader.ReadLine()
					if err != nil {
						logger.Error(fmt.Sprintf("error-reading-%s", driverFileName), err)
						return nil, err
					}
					address = string(addressBytes)
				case "json":
					// extract url from json file
					var driverJsonSpec voldriver.DriverSpec
					configFile, err := os.Open(path.Join(r.DriversPath, driverFileName))
					if err != nil {
						logger.Error(fmt.Sprintf("error-opening-config-%s", driverFileName), err)
						return nil, err
					}
					jsonParser := json.NewDecoder(configFile)
					if err = jsonParser.Decode(&driverJsonSpec); err != nil {
						logger.Error("parsing-config-file-error", err)
						return nil, err
					}
					address = driverJsonSpec.Address
				}

				logger.Info("invoking-driver", lager.Data{"address": address})
				driver, err = r.Factory.NewRemoteClient(address)
				if err != nil {
					logger.Error(fmt.Sprintf("error-building-driver-attached-to-%s", address), err)
					return nil, err
				}

				return driver, nil
			}
		}
	}
	return nil, fmt.Errorf("Driver '%s' not found in list of known drivers", driverId)
}

func (r *realDriverFactory) getMatchingDriverSpecs(logger lager.Logger, pattern string) ([]string, error) {
	matchingDriverSpecs, err := filepath.Glob(r.DriversPath + "/*." + pattern)
	if err != nil { // untestable on linux, does glob work differently on windows???
		return nil, fmt.Errorf("Volman configured with an invalid driver path '%s', error occured list files (%s)", r.DriversPath, err.Error())
	}
	logger.Info("binaries", lager.Data{"binaries": matchingDriverSpecs})
	return matchingDriverSpecs, nil

}
