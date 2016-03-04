package voldriver

import (
	"fmt"

	"github.com/cloudfoundry-incubator/volman/system"
	"github.com/pivotal-golang/lager"
)

type DriverClientCli struct {
	UseExec     system.Exec
	DriversPath string
	Name        string
}

func NewDriverClientCli(path string, useExec system.Exec, name string) *DriverClientCli {
	return &DriverClientCli{
		UseExec:     useExec,
		DriversPath: path,
		Name:        name,
	}
}

func (client *DriverClientCli) Mount(logger lager.Logger, mountRequest MountRequest) (MountResponse, error) {
	logger = logger.Session("driver-mount")
	logger.Info("start")
	defer logger.Info("end")
	response := struct {
		Path string `json:"path"`
	}{}
	invoker := NewCliInvoker(client.UseExec, fmt.Sprintf("%s/%s", client.DriversPath, client.Name), "mount", mountRequest.VolumeId, mountRequest.Config)
	err := invoker.InvokeDriver(logger, &response)
	if err != nil {
		return MountResponse{}, err
	}
	return MountResponse{response.Path}, nil
}

func (client *DriverClientCli) Unmount(logger lager.Logger, unmountRequest UnmountRequest) error {
	logger = logger.Session("driver-unmount")
	logger.Info("start")
	defer logger.Info("end")
	response := new(interface{})
	invoker := NewCliInvoker(client.UseExec, fmt.Sprintf("%s/%s", client.DriversPath, client.Name), "unmount", unmountRequest.VolumeId)
	return invoker.InvokeDriver(logger, &response)
}

func (client *DriverClientCli) Info(logger lager.Logger) (InfoResponse, error) {
	logger = logger.Session("driver-info")
	logger.Info("start")
	defer logger.Info("end")
	response := InfoResponse{}
	invoker := NewCliInvoker(client.UseExec, fmt.Sprintf("%s/%s", client.DriversPath, client.Name), "info")
	err := invoker.InvokeDriver(logger, &response)
	if err != nil {
		return InfoResponse{}, err
	}
	return response, nil
}
