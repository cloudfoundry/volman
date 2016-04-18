package vollocal

import (
	"sync"

	"fmt"

	"github.com/cloudfoundry-incubator/volman/voldriver"
)

type DriverRegistry interface {
	Driver(id string) (voldriver.Driver, bool)
	Activated(id string) (bool, error)
	Activate(id string) error
	Drivers() map[string]voldriver.Driver
	Add(id string, driver voldriver.Driver) error
	Keys() []string
}

type registryEntry struct {
	driver    voldriver.Driver
	activated bool
}

type driverRegistry struct {
	sync.RWMutex
	registryEntries map[string]*registryEntry
}

func NewDriverRegistry() DriverRegistry {
	return &driverRegistry{
		registryEntries: map[string]*registryEntry{},
	}
}

func NewDriverRegistryWith(initialMap map[string]voldriver.Driver) DriverRegistry {
	registryEntryMap := map[string]*registryEntry{}
	for name, driver := range initialMap {
		registryEntryMap[name] = &registryEntry{
			driver:    driver,
			activated: false,
		}
	}

	return &driverRegistry{
		registryEntries: registryEntryMap,
	}
}

func (d *driverRegistry) Driver(id string) (voldriver.Driver, bool) {
	d.RLock()
	defer d.RUnlock()

	if !d.containsDriver(id) {
		return nil, false
	}

	return d.registryEntries[id].driver, true
}

func (d *driverRegistry) Drivers() map[string]voldriver.Driver {
	d.RLock()
	defer d.RUnlock()

	driversCopy := map[string]voldriver.Driver{}
	for name, registryEntry := range d.registryEntries {
		driversCopy[name] = registryEntry.driver
	}

	return driversCopy
}

func (d *driverRegistry) Add(id string, driver voldriver.Driver) error {
	d.Lock()
	defer d.Unlock()

	if d.containsDriver(id) == false {
		d.registryEntries[id] = &registryEntry{
			driver:    driver,
			activated: false,
		}
		return nil
	}
	return fmt.Errorf("driver-exists")
}

func (d *driverRegistry) Keys() []string {
	d.Lock()
	defer d.Unlock()

	var keys []string
	for k := range d.registryEntries {
		keys = append(keys, k)
	}

	return keys
}

func (d *driverRegistry) Activated(id string) (bool, error) {
	d.Lock()
	defer d.Unlock()

	if !d.containsDriver(id) {
		return false, fmt.Errorf("driver-not-found")
	}

	driverEntry := d.registryEntries[id]

	return driverEntry.activated, nil
}

func (d *driverRegistry) Activate(id string) error {
	d.Lock()
	defer d.Unlock()

	if !d.containsDriver(id) {
		return fmt.Errorf("driver-not-found")
	}

	driverEntry := d.registryEntries[id]
	driverEntry.activated = true

	return nil
}

func (d *driverRegistry) containsDriver(id string) bool {
	_, ok := d.registryEntries[id]
	return ok
}
