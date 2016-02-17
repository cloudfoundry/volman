package delegate

import (
	"bytes"
	"encoding/json"
	"log"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry-incubator/volman"
	"github.com/pivotal-golang/lager"
)

type LocalClient struct {
	DriversPath string
}

func NewLocalClient(driversPath string) *LocalClient {
	return &LocalClient{driversPath}
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
		cmd := exec.Command(driverExecutable, "info")
		var out bytes.Buffer
		cmd.Stdout = &out

		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
			return volman.ListDriversResponse{}, err
		}

		var driver volman.Driver
		err = json.Unmarshal(out.Bytes(), &driver)
		if err != nil {
			log.Fatal(err)
			return volman.ListDriversResponse{}, err
		}

		response.Drivers = append(response.Drivers, driver)
	}
	return response, nil
}
