package delegate

import (
	"encoding/json"
	"fmt"
	"log"
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
	logger.Info("listing-drivers")
	driversBinaries, _ := filepath.Glob(client.DriversPath + "/*")

	if client.noDrivers(driversBinaries, client.DriversPath) {
		return volman.ListDriversResponse{}, nil
	}

	return client.listDrivers(driversBinaries)
}

func (client *LocalClient) noDrivers(driversBinaries []string, path string) bool {
	return len(driversBinaries) < 1 || path == ""
}

func (client *LocalClient) listDrivers(driversBinaries []string) (volman.ListDriversResponse, error) {
	var response volman.ListDriversResponse

	for _, driverExecutable := range driversBinaries {
		cmd := client.OurExec.Command(driverExecutable, "info")

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
			return volman.ListDriversResponse{}, err
		}
		if err := cmd.Start(); err != nil {
			log.Fatal(err)
			return volman.ListDriversResponse{}, err
		}
		var driver volman.Driver
		if err := json.NewDecoder(stdout).Decode(&driver); err != nil {
			return client.driverError(err, "decoding JSON", driverExecutable)
		}
		if err := cmd.Wait(); err != nil {
			log.Fatal(err)
			return volman.ListDriversResponse{}, err
		}

		response.Drivers = append(response.Drivers, driver)
	}
	return response, nil
}

func (client *LocalClient) driverError(err error, specifics string, driverExecutable string) (volman.ListDriversResponse, error) {
	log.Fatal(fmt.Sprintf("Error (%s) %s from binary at %s", err.Error(), specifics, driverExecutable))
	return volman.ListDriversResponse{}, err
}
