package vollocal

import (
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/volman"
	"github.com/tedsuo/ifrit"
	"os"
	"time"
)

type Syncer struct {
	logger       lager.Logger
	registry     volman.PluginRegistry
	scanInterval time.Duration
	clock        clock.Clock
	discoverer   []volman.Discoverer
}

func NewSyncer(logger lager.Logger, registry volman.PluginRegistry, discoverer []volman.Discoverer, scanInterval time.Duration, clock clock.Clock) *Syncer {
	return &Syncer{
		logger:       logger,
		registry:     registry,
		scanInterval: scanInterval,
		clock:        clock,
		discoverer:   discoverer,
	}
}

func NewSyncerWithShims(logger lager.Logger, registry volman.PluginRegistry, discoverer []volman.Discoverer, scanInterval time.Duration, clock clock.Clock) *Syncer {
	return &Syncer{
		logger:       logger,
		registry:     registry,
		scanInterval: scanInterval,
		clock:        clock,
		discoverer:   discoverer,
	}
}

func (p *Syncer) Runner() ifrit.Runner {
	return p
}

func (p *Syncer) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := p.logger.Session("sync-plugin")
	logger.Info("start")
	defer logger.Info("end")

	logger.Info("running-discovery")
	allPlugins := map[string]volman.Plugin{}
	for _, discoverer := range p.discoverer {
		plugins, err := discoverer.Discover(logger)
		if err != nil {
			logger.Error("failed-discover", err)
			return err
		}
		for k, v := range plugins {
			allPlugins[k] = v
		}
	}
	p.registry.Set(allPlugins)

	timer := p.clock.NewTimer(p.scanInterval)
	defer timer.Stop()

	close(ready)

	for {
		select {
		case <-timer.C():
			go func() {
				logger.Info("running-re-discovery")
				allPlugins := map[string]volman.Plugin{}
				for _, discoverer := range p.discoverer {
					plugins, err := discoverer.Discover(logger)
					if err != nil {
						logger.Error("failed-discover", err)
					}
					for k, v := range plugins {
						allPlugins[k] = v
					}
				}
				p.registry.Set(allPlugins)
				timer.Reset(p.scanInterval)
			}()
		case signal := <-signals:
			logger.Info("received-signal", lager.Data{"signal": signal.String()})
			return nil
		}
	}
	return nil
}
