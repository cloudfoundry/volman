package vollocal

import (
	"time"

	"sync"

	"os"

	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
)

type DriverSyncer interface {
	DriverRegistry() DriverRegistry

	Drivers() map[string]voldriver.Driver
	Driver(driverId string) (voldriver.Driver, bool)

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

func NewDriverSyncer(logger lager.Logger, driverFactory DriverFactory, scanInterval time.Duration, clock clock.Clock) *driverSyncer {
	return &driverSyncer{
		logger:        logger,
		driverFactory: driverFactory,
		scanInterval:  scanInterval,
		clock:         clock,

		driverRegistry: NewDriverRegistry(),
	}
}

func (d *driverSyncer) DriverRegistry() DriverRegistry {
	return d.driverRegistry
}

func (d *driverSyncer) Runner() ifrit.Runner {
	return d
}

func (d *driverSyncer) Drivers() map[string]voldriver.Driver {
	d.RLock()
	defer d.RUnlock()

	return d.driverRegistry.Drivers()
}

func (d *driverSyncer) Driver(driverId string) (voldriver.Driver, bool) {
	d.RLock()
	defer d.RUnlock()

	return d.driverRegistry.Driver(driverId)
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
