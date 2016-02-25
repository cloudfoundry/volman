package driverclient

import (
	"fmt"

	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/system"
	"github.com/pivotal-golang/lager"
)

type Mounting struct {
	Driver     volman.Driver
	DriverList []volman.Driver
	UseExec    system.Exec
}

func (mounting *Mounting) Mount(logger lager.Logger, volumeId string, config string) (string, error) {
	driverExecutable, err := mounting.driverExecutable(logger)
	if err != nil {
		return "", err
	}

	response := struct {
		Path string `json:"path"`
	}{}

	invoker := NewCliInvoker(mounting.UseExec, driverExecutable, "mount", volumeId, config)
	err = invoker.InvokeDriver(logger, &response)
	if err != nil {
		return "", err
	}

	return response.Path, nil
}

func (mounting *Mounting) driverExecutable(logger lager.Logger) (string, error) {
	var driverExecutable string
	for _, potentialDriver := range mounting.DriverList {
		if potentialDriver.Name == mounting.Driver.Name {
			driverExecutable = potentialDriver.Binary
		}
	}
	if driverExecutable == "" {
		return "", mounting.logError(logger, fmt.Errorf("Driver '%s' not found in list of known drivers", mounting.Driver.Name), "Error finding driver", mounting.Driver.Name)
	}

	return driverExecutable, nil
}

func (mounting *Mounting) logError(logger lager.Logger, err error, specifics string, name string) error {
	logger.Error(fmt.Sprintf("Error (%s) %s from driver at %s", err.Error(), specifics, name), err)
	return err
}
