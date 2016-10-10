package vollocal

import (
	"errors"
	"os"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/voldriver"
	"github.com/tedsuo/ifrit"
)

type MountPurger interface {
	Runner() ifrit.Runner
	PurgeMounts(logger lager.Logger) error
}

type mountPurger struct {
	logger   lager.Logger
	registry DriverRegistry
}

func NewMountPurger(logger lager.Logger, registry DriverRegistry) MountPurger {
	return &mountPurger{
		logger,
		registry,
	}
}

func (p *mountPurger) Runner() ifrit.Runner {
	return p
}

func (p *mountPurger) Run(signals <-chan os.Signal, ready chan<- struct{}) error {

	if err := p.PurgeMounts(p.logger); err != nil {
		return err
	}

	close(ready)
	<-signals
	return nil
}

func (p *mountPurger) PurgeMounts(logger lager.Logger) error {
	logger = logger.Session("purge-mounts")
	logger.Info("start")
	defer logger.Info("end")

	drivers := p.registry.Drivers()

	for _, driver := range drivers {
		listResponse := driver.List(logger)
		for _, mount := range listResponse.Volumes {
			errorResponse := driver.Unmount(logger, voldriver.UnmountRequest{Name: mount.Name})
			if errorResponse.Err != "" {
				logger.Error("failed-purging-volume-mount", errors.New(errorResponse.Err))
			}
		}
	}

	return nil
}
