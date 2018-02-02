package vollocal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"errors"
	"context"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/voldriver"
	"code.cloudfoundry.org/voldriver/driverhttp"
	"code.cloudfoundry.org/volman"
)

type dockerDriverDiscoverer struct {
	logger        lager.Logger
	driverFactory DockerDriverFactory

	driverRegistry volman.PluginRegistry
	driverPaths    []string
}

func NewDockerDriverDiscoverer(logger lager.Logger, driverRegistry volman.PluginRegistry, driverPaths []string) volman.Discoverer {
	return &dockerDriverDiscoverer{
		logger:        logger,
		driverFactory: NewDockerDriverFactory(),

		driverRegistry: driverRegistry,
		driverPaths:    driverPaths,
	}
}

func NewDockerDriverDiscovererWithDriverFactory(logger lager.Logger, driverRegistry volman.PluginRegistry, driverPaths []string, factory DockerDriverFactory) volman.Discoverer {
	return &dockerDriverDiscoverer{
		logger:        logger,
		driverFactory: factory,

		driverRegistry: driverRegistry,
		driverPaths:    driverPaths,
	}
}

func (r *dockerDriverDiscoverer) Discover(logger lager.Logger) (map[string]volman.Plugin, error) {
	logger = logger.Session("discover")
	logger.Debug("start")
	logger.Info("discovering-drivers", lager.Data{"driver-paths": r.driverPaths})
	defer logger.Debug("end")

	endpoints := make(map[string]volman.Plugin)

	for _, driverPath := range r.driverPaths {
		//precedence order: sock -> spec -> json
		spec_types := [3]string{"sock", "spec", "json"}
		for _, spec_type := range spec_types {
			matchingDriverSpecs, err := r.getMatchingDriverSpecs(logger, driverPath, spec_type)

			if err != nil {
				// untestable on linux, does glob work differently on windows???
				return map[string]volman.Plugin{}, fmt.Errorf("Volman configured with an invalid driver path '%s', error occured list files (%s)", driverPath, err.Error())
			}
			if len(matchingDriverSpecs) > 0 {
				logger.Debug("driver-specs", lager.Data{"drivers": matchingDriverSpecs})
				var existing map[string]volman.Plugin
				if r.driverRegistry != nil {
					existing = r.driverRegistry.Plugins()
					logger.Debug("existing-drivers", lager.Data{"len": len(existing)})
				}

				endpoints = r.insertIfAliveAndNotFound(logger, endpoints, driverPath, matchingDriverSpecs, existing)
			}
		}
	}
	return endpoints, nil
}

func (r *dockerDriverDiscoverer) getMatchingDriverSpecs(logger lager.Logger, path string, pattern string) ([]string, error) {
	logger.Debug("binaries", lager.Data{"path": path, "pattern": pattern})
	matchingDriverSpecs, err := filepath.Glob(path + string(os.PathSeparator) + "*." + pattern)
	if err != nil { // untestable on linux, does glob work differently on windows???
		return nil, fmt.Errorf("Volman configured with an invalid driver path '%s', error occured list files (%s)", path, err.Error())
	}
	return matchingDriverSpecs, nil

}

func (r *dockerDriverDiscoverer) insertIfAliveAndNotFound(logger lager.Logger, endpoints map[string]volman.Plugin, driverPath string, specs []string, existing map[string]volman.Plugin) map[string]volman.Plugin {
	logger = logger.Session("insert-if-not-found")
	logger.Debug("start")
	defer logger.Debug("end")

	var plugin volman.Plugin
	var ok bool
	var re *regexp.Regexp

	for _, spec := range specs {
		if runtime.GOOS == "windows" {
			re = regexp.MustCompile(`([^\\]*\\)?([^\\]*)\.(sock|spec|json)$`)
		} else {
			re = regexp.MustCompile(`([^/]*/)?([^/]*)\.(sock|spec|json)$`)
		}

		segs2 := re.FindAllStringSubmatch(spec, 1)
		if len(segs2) <= 0 {
			continue
		}
		specName := segs2[0][2]
		specFile := segs2[0][2] + "." + segs2[0][3]
		logger.Debug("insert-unique-spec", lager.Data{"specname": specName})

		_, ok = endpoints[specName]
		if !ok {
			plugin, ok = existing[specName]
			if ok == true {
				driverSpec, err := voldriver.ReadDriverSpec(logger, specName, driverPath, specFile)
				if err != nil {
					logger.Error("error-reading-driver-spec", err)
					continue
				}
				pluginSpec := volman.PluginSpec{
					Name:    driverSpec.Name,
					Address: driverSpec.Address,
				}
				if driverSpec.TLSConfig != nil {
					pluginSpec.TLSConfig = &volman.TLSConfig{
						InsecureSkipVerify: driverSpec.TLSConfig.InsecureSkipVerify,
						CAFile:             driverSpec.TLSConfig.CAFile,
						CertFile:           driverSpec.TLSConfig.CertFile,
						KeyFile:            driverSpec.TLSConfig.KeyFile,
					}
				}
				if !plugin.Matches(logger, pluginSpec) {
					logger.Info("existing-driver-mismatch", lager.Data{"specName": specName, "address": driverSpec.Address, "tls": driverSpec.TLSConfig})
					plugin = nil
				}
				if plugin != nil {
					dockerPlugin := plugin.(*driverhttp.DockerDriverPlugin)
					dockerDriver := dockerPlugin.DockerDriver.(voldriver.Driver)
					env := driverhttp.NewHttpDriverEnv(logger, context.Background())
					resp := dockerDriver.Activate(env)
					if resp.Err != "" {
						logger.Error("existing-driver-unreachable", errors.New(resp.Err), lager.Data{"specName": specName, "address": driverSpec.Address, "tls": driverSpec.TLSConfig})
						plugin = nil
					}
				}
			}

			if plugin == nil {
				logger.Info("creating-driver", lager.Data{"specName": specName, "driver-path": driverPath, "specFile": specFile})
				driver, err := r.driverFactory.DockerDriver(logger, specName, driverPath, specFile)
				if err != nil {
					logger.Error("error-creating-driver", err)
					continue
				}

				plugin = driverhttp.NewDockerPluginWithDriver(driver)

				env := driverhttp.NewHttpDriverEnv(logger, context.TODO())
				resp := driver.Activate(env)
				if resp.Err != "" {
					logger.Info("skipping-non-responsive-driver", lager.Data{"specname": specName})
					continue
				} else {
					driverImplementsErr := fmt.Errorf("driver-implements: %#v", resp.Implements)
					if len(resp.Implements) == 0 {
						logger.Error("driver-incorrect", driverImplementsErr)
						continue
					}

					if !driverImplements("VolumeDriver", resp.Implements) {
						logger.Error("driver-incorrect", driverImplementsErr)
						continue
					}
				}
			}
			logger.Info("new-driver", lager.Data{"name": specName})
			endpoints[specName] = plugin
		}
	}
	return endpoints
}
