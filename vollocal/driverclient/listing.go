package driverclient

import (
	"fmt"

	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/system"
	"github.com/pivotal-golang/lager"
)

type Listing struct {
	DriversBinaries []string
	UseExec         system.Exec
}

func (listing *Listing) List(logger lager.Logger, path string) ([]volman.Driver, error) {
	if listing.noDrivers(path) {
		return nil, nil
	}
	return listing.listDrivers(logger)
}

func (listing *Listing) noDrivers(path string) bool {
	return len(listing.DriversBinaries) < 1 || path == ""
}

func (listing *Listing) listDrivers(logger lager.Logger) ([]volman.Driver, error) {
	var response []volman.Driver

	for _, driverExecutable := range listing.DriversBinaries {

		var driver volman.Driver

		invoker := NewCliInvoker(listing.UseExec, driverExecutable, "info")
		err := invoker.InvokeDriver(logger, &driver)
		if err != nil {
			logger.Error("Failure invoking driver to get info", err)
			return nil, err
		}

		driver.Binary = driverExecutable
		logger.Info(fmt.Sprintf("Found driver %#v", driver))
		response = append(response, driver)
	}
	return response, nil
}
