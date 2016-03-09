package voldriver

import (
	"encoding/json"
	"os"

	"github.com/pivotal-golang/lager"
)

func WriteDriverSpec(logger lager.Logger, pluginsDirectory string, driver string, url string) error {
	spec := DriverSpec{
		Name:    driver,
		Address: url,
	}

	return WriteDriverSpecWithSpec(logger, pluginsDirectory, driver, spec)
}

func WriteDriverSpecWithSpec(logger lager.Logger, pluginsDirectory string, driver string, spec DriverSpec) error {

	contents, err := json.Marshal(spec)
	if err != nil {
		logger.Error("Error writing to file "+err.Error(), err)
		return err
	}

	return WriteDriverSpecWithContents(logger, pluginsDirectory, driver, contents)
}

func WriteDriverSpecWithContents(logger lager.Logger, pluginsDirectory string, driver string, contents []byte) error {
	f, err := os.Create(pluginsDirectory + "/" + driver + ".json")
	if err != nil {
		logger.Error("Error creating file "+err.Error(), err)
		return err
	}
	defer f.Close()

	_, err = f.Write(contents)
	if err != nil {
		logger.Error("Error writing to file "+err.Error(), err)
		return err
	}
	f.Sync()
	return nil
}
