package voldriver

import (
	"encoding/json"

	"github.com/pivotal-golang/lager"
)

type specWriter struct {
	driverName string
	path       string

	logger lager.Logger
}

type SpecWriter interface {
	WriteJson(driverSpec DriverSpec) error
	WriteSpec(address string) error
}

func NewSpecWriter(logger lager.Logger, driverName string, path string) specWriter {
	return specWriter{
		driverName: driverName,
		path:       path,

		logger: logger,
	}
}

func (sw specWriter) WriteJson(driverSpec DriverSpec) error {
	contents, err := json.Marshal(driverSpec)
	if err != nil {
		return err
	}

	return WriteDriverSpec(sw.logger, sw.path, sw.driverName, "json", contents)
}

func (sw specWriter) WriteSpec(address string) error {
	return WriteDriverSpec(sw.logger, sw.path, sw.driverName, "spec", []byte(address))
}
