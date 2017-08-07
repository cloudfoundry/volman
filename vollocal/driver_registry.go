package vollocal

import (
	"sync"

	"code.cloudfoundry.org/voldriver"
)

type DriverRegistry interface {
	Driver(id string) (voldriver.Plugin, bool)
	Drivers() map[string]voldriver.Plugin
	Set(drivers map[string]voldriver.Plugin)
	Keys() []string
}

type driverRegistry struct {
	sync.RWMutex
	registryEntries map[string]voldriver.Plugin
}

func NewDriverRegistry() DriverRegistry {
	return &driverRegistry{
		registryEntries: map[string]voldriver.Plugin{},
	}
}

func NewDriverRegistryWith(initialMap map[string]voldriver.Plugin) DriverRegistry {
	return &driverRegistry{
		registryEntries: initialMap,
	}
}

func (d *driverRegistry) Driver(id string) (voldriver.Plugin, bool) {
	d.RLock()
	defer d.RUnlock()

	if !d.containsDriver(id) {
		return nil, false
	}

	return d.registryEntries[id], true
}

func (d *driverRegistry) Drivers() map[string]voldriver.Plugin {
	d.RLock()
	defer d.RUnlock()

	return d.registryEntries
}

func (d *driverRegistry) Set(drivers map[string]voldriver.Plugin) {
	d.Lock()
	defer d.Unlock()

	d.registryEntries = drivers
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

func (d *driverRegistry) containsDriver(id string) bool {
	_, ok := d.registryEntries[id]
	return ok
}
