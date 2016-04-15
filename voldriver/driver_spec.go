package voldriver

import (
	"os"

	"github.com/pivotal-golang/lager"
)

func WriteDriverSpec(logger lager.Logger, pluginsDirectory string, driver string, url string) error {
	os.MkdirAll(pluginsDirectory, 0666)
	f, err := os.Create(pluginsDirectory + "/" + driver + ".spec")
	if err != nil {
		logger.Error("Error creating file "+err.Error(), err)
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(url))
	if err != nil {
		logger.Error("Error writing to file "+err.Error(), err)
		return err
	}
	f.Sync()
	return nil
}

func WriteDriverJSONSpec(logger lager.Logger, pluginsDirectory string, driver string, contents []byte) error {
	os.MkdirAll(pluginsDirectory, 0666)
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
