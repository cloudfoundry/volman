package vollocal

import (
	"time"

	"sync"

	"os"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-incubator/voldriver"
	"github.com/tedsuo/ifrit"
)

type DriverSyncer interface {
	Runner() ifrit.Runner
}

type driverSyncer struct {
	sync.RWMutex
	logger        lager.Logger
	driverFactory DriverFactory
	scanInterval  time.Duration
	clock         clock.Clock

	driverRegistry DriverRegistry
}

func NewDriverSyncer(logger lager.Logger, driverRegistry DriverRegistry, driverPaths []string, scanInterval time.Duration, clock clock.Clock) *driverSyncer {
	return &driverSyncer{
		logger:        logger,
		driverFactory: NewDriverFactory(driverPaths),
		scanInterval:  scanInterval,
		clock:         clock,

		driverRegistry: driverRegistry,
	}
}

func NewDriverSyncerWithDriverFactory(logger lager.Logger, driverRegistry DriverRegistry, driverPaths []string, scanInterval time.Duration, clock clock.Clock, factory DriverFactory) *driverSyncer {
	return &driverSyncer{
		logger:        logger,
		driverFactory: factory,
		scanInterval:  scanInterval,
		clock:         clock,

		driverRegistry: driverRegistry,
	}
}

func (d *driverSyncer) Runner() ifrit.Runner {
	return d
}

func (r *driverSyncer) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := r.logger.Session("sync-drivers")
	logger.Info("start")
	defer logger.Info("end")

	timer := r.clock.NewTimer(r.scanInterval)
	defer timer.Stop()

	drivers, err := r.driverFactory.Discover(logger)
	if err != nil {
		return err
	}

	r.addNewDrivers(drivers)

	close(ready)

	setDriverCh := make(chan error, 1)

	for {
		select {

		case <-setDriverCh:
			timer.Reset(r.scanInterval)

		case <-timer.C():
			drivers, err := r.driverFactory.Discover(logger)
			if err != nil {
				setDriverCh <- err
				continue
			}

			r.addNewDrivers(drivers)
			setDriverCh <- nil

		case signal := <-signals:
			logger.Info("received-signal", lager.Data{"signal": signal.String()})
			return nil
		}
	}
}

func (r *driverSyncer) addNewDrivers(drivers map[string]voldriver.Driver) {
	for name, driver := range drivers {
		if _, exists := r.driverRegistry.Driver(name); exists == false {
			r.driverRegistry.Add(name, driver)
		}
	}
}
