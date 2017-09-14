package voldiscoverers

import (
	"path/filepath"
	"code.cloudfoundry.org/goshims/filepathshim"
	"code.cloudfoundry.org/goshims/grpcshim"
	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/volman"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"code.cloudfoundry.org/csishim"
	. "github.com/Kaixiang/csiplugin"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type csiPluginDiscoverer struct {
	logger          lager.Logger
	pluginRegistry  volman.PluginRegistry
	pluginPaths     []string
	filepathShim    filepathshim.Filepath
	grpcShim        grpcshim.Grpc
	csiShim         csishim.Csi
	osShim          osshim.Os
	csiMountRootDir string
}

func NewCsiPluginDiscoverer(logger lager.Logger, pluginRegistry volman.PluginRegistry, pluginPaths []string, csiMountRootDir string) volman.Discoverer {
	return &csiPluginDiscoverer{
		logger:          logger,
		pluginRegistry:  pluginRegistry,
		pluginPaths:     pluginPaths,
		filepathShim:    &filepathshim.FilepathShim{},
		grpcShim:        &grpcshim.GrpcShim{},
		csiShim:         &csishim.CsiShim{},
		osShim:          &osshim.OsShim{},
		csiMountRootDir: csiMountRootDir,
	}
}

func NewCsiPluginDiscovererWithShims(logger lager.Logger, pluginRegistry volman.PluginRegistry, pluginPaths []string, filepathShim filepathshim.Filepath, grpcShim grpcshim.Grpc, csiShim csishim.Csi, osShim osshim.Os, csiMountRootDir string) volman.Discoverer {
	return &csiPluginDiscoverer{
		logger:          logger,
		pluginRegistry:  pluginRegistry,
		pluginPaths:     pluginPaths,
		filepathShim:    filepathShim,
		grpcShim:        grpcShim,
		csiShim:         csiShim,
		osShim:          osShim,
		csiMountRootDir: csiMountRootDir,
	}
}

func (p *csiPluginDiscoverer) Discover(logger lager.Logger) (map[string]volman.Plugin, error) {
	logger = logger.Session("discover")
	logger.Debug("start")
	defer logger.Debug("end")
	var conns []grpcshim.ClientConn
	defer func() {
		for _, conn := range conns {
			err := conn.Close()
			if err != nil {
				logger.Error("grpc-conn-close", err)
			}
		}
	}()

	plugins := map[string]volman.Plugin{}

	for _, pluginPath := range p.pluginPaths {
		pluginSpecFiles, err := filepath.Glob(pluginPath + "/*.json")
		if err != nil {
			logger.Error("filepath-glob", err, lager.Data{"glob": pluginPath + "/*.json"})
			return plugins, err
		}
		for _, pluginSpecFile := range pluginSpecFiles {
			csiPluginSpec, err := ReadSpec(logger, pluginSpecFile)
			if err != nil {
				logger.Error("read-spec-failed", err, lager.Data{"plugin-path": pluginPath, "plugin-spec-file": pluginSpecFile})
				continue
			}

			existingPlugin, found := p.pluginRegistry.Plugins()[csiPluginSpec.Name]
			pluginSpec := volman.PluginSpec{
				Name:    csiPluginSpec.Name,
				Address: csiPluginSpec.Address,
			}

			if !found || !existingPlugin.Matches(logger, pluginSpec) {
				logger.Info("new-plugin", lager.Data{"name": pluginSpec.Name, "address": pluginSpec.Address})

				// instantiate a volman.Plugin implementation of a csi.NodePlugin
				conn, err := p.grpcShim.Dial(csiPluginSpec.Address, grpc.WithInsecure())
				conns = append(conns, conn)
				if err != nil {
					logger.Error("grpc-dial", err, lager.Data{"address": csiPluginSpec.Address})
					continue
				}

				nodePlugin := p.csiShim.NewNodeClient(conn)
				_, err = nodePlugin.ProbeNode(context.TODO(), &csi.ProbeNodeRequest{
					Version: &csi.Version{
						Major: 0,
						Minor: 0,
						Patch: 1,
					},
				})
				if err != nil {
					logger.Info("probe-node-unresponsive", lager.Data{"name": csiPluginSpec.Name, "address": csiPluginSpec.Address})
					continue
				}

				plugin := NewCsiPlugin(nodePlugin, pluginSpec, p.grpcShim, p.csiShim, p.osShim, p.csiMountRootDir)
				plugins[csiPluginSpec.Name] = plugin
			} else {
				logger.Info("discovered-plugin-ignored", lager.Data{"name": pluginSpec.Name, "address": pluginSpec.Address})
				plugins[csiPluginSpec.Name] = existingPlugin
			}
		}
	}
	return plugins, nil
}
