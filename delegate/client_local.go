package delegate

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/system"
	"github.com/pivotal-golang/lager"
)

type LocalClient struct {
	DriversPath string
	OurExec     system.Exec
}

func NewLocalClient(driversPath string) *LocalClient {
	return &LocalClient{driversPath, &system.SystemExec{}}
}

func (client *LocalClient) ListDrivers(logger lager.Logger) (volman.ListDriversResponse, error) {
	logger = logger.Session("list-drivers")
	logger.Info("start")
	driversBinaries, _ := filepath.Glob(client.DriversPath + "/*")

	if client.noDrivers(driversBinaries, client.DriversPath) {
		return volman.ListDriversResponse{}, nil
	}

	defer logger.Info("end")
	return client.listDrivers(logger, driversBinaries)
}

func (client *LocalClient) noDrivers(driversBinaries []string, path string) bool {
	return len(driversBinaries) < 1 || path == ""
}

func (client *LocalClient) listDrivers(logger lager.Logger, driversBinaries []string) (volman.ListDriversResponse, error) {
	var response volman.ListDriversResponse

	for _, driverExecutable := range driversBinaries {
		cmd := client.OurExec.Command(driverExecutable, "info")

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return client.driverError(logger, err, "fetching stdout", driverExecutable)
		}
		if err := cmd.Start(); err != nil {
			return client.driverError(logger, err, "starting", driverExecutable)
		}
		var driver volman.Driver
		if err := json.NewDecoder(stdout).Decode(&driver); err != nil {
			return client.driverError(logger, err, "decoding JSON", driverExecutable)
		}
		if err := cmd.Wait(); err != nil {
			return client.driverError(logger, err, "waiting", driverExecutable)
		}

		response.Drivers = append(response.Drivers, driver)
	}
	return response, nil
}

func (client *LocalClient) driverError(logger lager.Logger, err error, specifics string, driverExecutable string) (volman.ListDriversResponse, error) {
	logger.Error(fmt.Sprintf("Error (%s) %s from driver at %s", err.Error(), specifics, driverExecutable), err)
	return volman.ListDriversResponse{}, err
}
