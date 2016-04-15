package vollocal

import (
	"time"

	"sync"

	"os"

	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
)

type driverMap struct {
	sync.RWMutex
	drivers map[string]voldriver.Driver
}

func (d *driverMap) Driver(id string) (voldriver.Driver, bool) {
	d.RLock()
	defer d.RUnlock()
	_, ok := d.drivers[id]
	if !ok {
		return nil, false
	}

	return d.drivers[id], true

}

func (d *driverMap) Drivers() map[string]voldriver.Driver {
	d.RLock()
	defer d.RUnlock()

	return d.drivers
}

func (d *driverMap) Set(drivers map[string]voldriver.Driver) {
	d.Lock()
	defer d.Unlock()

	d.drivers = drivers
}

func (d *driverMap) Keys() []string {
	d.Lock()
	defer d.Unlock()

	var keys []string
	for k := range d.drivers {
		keys = append(keys, k)
	}

	return keys
}

type DriverSyncer struct {
	sync.RWMutex
	logger        lager.Logger
	driverFactory DriverFactory
	scanInterval  time.Duration
	clock         clock.Clock

	drivers driverMap
}

func NewDriverSyncer(logger lager.Logger, driverFactory DriverFactory, scanInterval time.Duration, clock clock.Clock) *DriverSyncer {
	return &DriverSyncer{
		logger:        logger,
		driverFactory: driverFactory,
		scanInterval:  scanInterval,
		clock:         clock,

		drivers: driverMap{},
	}
}

func (d *DriverSyncer) Drivers() map[string]voldriver.Driver {
	d.RLock()
	defer d.RUnlock()

	return d.drivers.Drivers()
}

func (d *DriverSyncer) Driver(driverId string) (voldriver.Driver, bool) {
	d.RLock()
	defer d.RUnlock()

	return d.drivers.Driver(driverId)
}

func (r *DriverSyncer) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := r.logger.Session("sync-drivers")
	logger.Info("start")
	defer logger.Info("end")

	timer := r.clock.NewTimer(r.scanInterval)
	defer timer.Stop()

	drivers, err := r.driverFactory.Discover(logger)
	if err != nil {
		return err
	}

	r.drivers.Set(drivers)

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

			r.drivers.Set(drivers)
			setDriverCh <- nil

		case signal := <-signals:
			logger.Info("received-signal", lager.Data{"signal": signal.String()})
			return nil
		}
	}
}
