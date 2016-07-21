package vollocal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"net/url"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-incubator/voldriver"
	"github.com/cloudfoundry-incubator/voldriver/driverhttp"
	"github.com/cloudfoundry/gunk/os_wrap"
)

//go:generate counterfeiter -o ../volmanfakes/fake_driver_factory.go . DriverFactory

// DriverFactories are responsible for instantiating remote client implementations of the voldriver.Driver interface.
type DriverFactory interface {

	// Discover will compile a list of drivers from the path list in DriversPath.  If the same driver id is found in
	// multiple directories, it will favor the directory found first in the path.
	// if 2 specs are found within the *same* directory, it will choose .sock files first, then .spec files, then .json
	Discover(logger lager.Logger) (map[string]voldriver.Driver, error)

	// Given a driver id, path and config filename returns a remote client implementation of the voldriver.Driver interface
	Driver(logger lager.Logger, driverId string, driverPath, driverFileName string) (voldriver.Driver, error)
}

type realDriverFactory struct {
	DriverPaths     []string
	Factory         driverhttp.RemoteClientFactory
	useOs           os_wrap.Os
	DriversRegistry map[string]voldriver.Driver
}

func NewDriverFactory(driverPaths []string) DriverFactory {
	remoteClientFactory := driverhttp.NewRemoteClientFactory()
	return NewDriverFactoryWithRemoteClientFactory(driverPaths, remoteClientFactory)
}

func NewDriverFactoryWithRemoteClientFactory(driverPaths []string, remoteClientFactory driverhttp.RemoteClientFactory) DriverFactory {
	return &realDriverFactory{driverPaths, remoteClientFactory, os_wrap.NewOs(), nil}
}

func NewDriverFactoryWithOs(driverPaths []string, useOs os_wrap.Os) DriverFactory {
	remoteClientFactory := driverhttp.NewRemoteClientFactory()
	return &realDriverFactory{driverPaths, remoteClientFactory, useOs, nil}
}

func (r *realDriverFactory) Discover(logger lager.Logger) (map[string]voldriver.Driver, error) {
	logger = logger.Session("discover")
	logger.Debug("start")
	logger.Info(fmt.Sprintf("Discovering drivers in %s", r.DriverPaths))
	defer logger.Debug("end")

	endpoints := make(map[string]voldriver.Driver)
	for _, driverPath := range r.DriverPaths {
		//precedence order: sock -> spec -> json
		spec_types := [3]string{"sock", "spec", "json"}
		for _, spec_type := range spec_types {
			matchingDriverSpecs, err := r.getMatchingDriverSpecs(logger, driverPath, spec_type)

			if err != nil {
				// untestable on linux, does glob work differently on windows???
				return map[string]voldriver.Driver{}, fmt.Errorf("Volman configured with an invalid driver path '%s', error occured list files (%s)", driverPath, err.Error())
			}
			if len(matchingDriverSpecs) > 0 {
				logger.Debug("driver-specs", lager.Data{"drivers": matchingDriverSpecs})
				endpoints = r.insertIfAliveAndNotFound(logger, endpoints, driverPath, matchingDriverSpecs)
			}
		}
	}
	return endpoints, nil
}

func (r *realDriverFactory) insertIfAliveAndNotFound(logger lager.Logger, endpoints map[string]voldriver.Driver, driverPath string, specs []string) map[string]voldriver.Driver {
	logger = logger.Session("insert-if-not-found")
	logger.Debug("start")
	defer logger.Debug("end")

	for _, spec := range specs {
		re := regexp.MustCompile("([^/]*/)?([^/]*)\\.(sock|spec|json)$")

		segs2 := re.FindAllStringSubmatch(spec, 1)
		if len(segs2) <= 0 {
			continue
		}
		specName := segs2[0][2]
		specFile := segs2[0][2] + "." + segs2[0][3]
		logger.Debug("insert-unique-spec", lager.Data{"specname": specName})
		_, ok := endpoints[specName]
		if ok == false {
			driver, err := r.Driver(logger, specName, driverPath, specFile)
			if err != nil {
				logger.Error("error-creating-driver", err)
				continue
			}

			resp := driver.List(logger)
			if resp.Err != "" {
				logger.Info("skipping-non-responsive-driver", lager.Data{"specname": specName})
			} else {
				endpoints[specName] = driver
			}
		}
	}
	return endpoints
}

func (r *realDriverFactory) Driver(logger lager.Logger, driverId string, driverPath string, driverFileName string) (voldriver.Driver, error) {
	logger = logger.Session("driver", lager.Data{"driverId": driverId, "driverFileName": driverFileName})
	logger.Info("start")
	defer logger.Info("end")

	var driver voldriver.Driver

	var address string
	var tls *voldriver.TLSConfig
	if strings.Contains(driverFileName, ".") {
		extension := strings.Split(driverFileName, ".")[1]
		switch extension {
		case "sock":
			address = path.Join(driverPath, driverFileName)
		case "spec":
			configFile, err := r.useOs.Open(path.Join(driverPath, driverFileName))
			if err != nil {
				logger.Error(fmt.Sprintf("error-opening-config-%s", driverFileName), err)
				return nil, err
			}
			reader := bufio.NewReader(configFile)
			addressBytes, _, err := reader.ReadLine()
			if err != nil { // no real value in faking this as bigger problems exist when this fails
				logger.Error(fmt.Sprintf("error-reading-%s", driverFileName), err)
				return nil, err
			}
			address = string(addressBytes)
		case "json":
			// extract url from json file
			var driverJsonSpec voldriver.DriverSpec
			configFile, err := r.useOs.Open(path.Join(driverPath, driverFileName))
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
			tls = driverJsonSpec.TLSConfig
		default:
			err := fmt.Errorf("unknown-driver-extension: %s", extension)
			logger.Error("driver", err)
			return nil, err

		}
		var err error

		address, err = r.canonicalize(logger, address)
		if err != nil {
			logger.Error(fmt.Sprintf("invalid-address: %s", address), err)
			return nil, err
		}

		logger.Info("getting-driver", lager.Data{"address": address})
		driver, err = r.Factory.NewRemoteClient(address, tls)
		if err != nil {
			logger.Error(fmt.Sprintf("error-building-driver-attached-to-%s", address), err)
			return nil, err
		}

		return driver, nil
	}

	return nil, fmt.Errorf("Driver '%s' not found in list of known drivers", driverId)
}

func (r *realDriverFactory) getMatchingDriverSpecs(logger lager.Logger, path string, pattern string) ([]string, error) {
	logger.Debug("binaries", lager.Data{"path": path, "pattern": pattern})
	matchingDriverSpecs, err := filepath.Glob(path + "/*." + pattern)
	if err != nil { // untestable on linux, does glob work differently on windows???
		return nil, fmt.Errorf("Volman configured with an invalid driver path '%s', error occured list files (%s)", path, err.Error())
	}
	return matchingDriverSpecs, nil

}

func (r *realDriverFactory) canonicalize(logger lager.Logger, address string) (string, error) {
	logger = logger.Session("canonicalize", lager.Data{"address": address})
	logger.Debug("start")
	defer logger.Debug("end")

	url, err := url.Parse(address)
	if err != nil {
		return address, err
	}

	switch url.Scheme {
	case "http", "https":
		return address, nil
	case "tcp":
		return fmt.Sprintf("http://%s%s", url.Host, url.Path), nil
	case "unix":
		return address, nil
	default:
		if strings.HasSuffix(url.Path, ".sock") {
			return fmt.Sprintf("%s%s", url.Host, url.Path), nil
		}
	}
	return fmt.Sprintf("http://%s", address), nil
}
