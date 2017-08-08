package vollocal

import (
	"sync"

	"code.cloudfoundry.org/voldriver"
)

type PluginRegistry interface {
	Plugin(id string) (voldriver.Plugin, bool)
	Plugins() map[string]voldriver.Plugin
	Set(plugins map[string]voldriver.Plugin)
	Keys() []string
}

type pluginRegistry struct {
	sync.RWMutex
	registryEntries map[string]voldriver.Plugin
}

func NewPluginRegistry() PluginRegistry {
	return &pluginRegistry{
		registryEntries: map[string]voldriver.Plugin{},
	}
}

func NewPluginRegistryWith(initialMap map[string]voldriver.Plugin) PluginRegistry {
	return &pluginRegistry{
		registryEntries: initialMap,
	}
}

func (d *pluginRegistry) Plugin(id string) (voldriver.Plugin, bool) {
	d.RLock()
	defer d.RUnlock()

	if !d.containsPlugin(id) {
		return nil, false
	}

	return d.registryEntries[id], true
}

func (d *pluginRegistry) Plugins() map[string]voldriver.Plugin {
	d.RLock()
	defer d.RUnlock()

	return d.registryEntries
}

func (d *pluginRegistry) Set(plugins map[string]voldriver.Plugin) {
	d.Lock()
	defer d.Unlock()

	d.registryEntries = plugins
}

func (d *pluginRegistry) Keys() []string {
	d.Lock()
	defer d.Unlock()

	var keys []string
	for k := range d.registryEntries {
		keys = append(keys, k)
	}

	return keys
}

func (d *pluginRegistry) containsPlugin(id string) bool {
	_, ok := d.registryEntries[id]
	return ok
}
