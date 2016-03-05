package voldriver

import (
	"fmt"

	"github.com/cloudfoundry-incubator/volman/system"
	"github.com/pivotal-golang/lager"
)

type DriverClientCli struct {
	Invoker *CliInvoker
}

func NewDriverClientCli(path string, useExec system.Exec, name string) *DriverClientCli {
	return &DriverClientCli{
		NewCliInvoker(useExec, fmt.Sprintf("%s/%s", path, name))}
}

func (client *DriverClientCli) Mount(logger lager.Logger, mountRequest MountRequest) (MountResponse, error) {
	logger = logger.Session("driver-mount")
	logger.Info("start")
	defer logger.Info("end")
	response := struct {
		Path string `json:"path"`
	}{}
	client.Invoker.Command("mount", mountRequest.VolumeId, mountRequest.Config)
	err := client.Invoker.Execute(logger, &response)
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
	client.Invoker.Command("unmount", unmountRequest.VolumeId)
	return client.Invoker.Execute(logger, &response)
}

func (client *DriverClientCli) Info(logger lager.Logger) (InfoResponse, error) {
	logger = logger.Session("driver-info")
	logger.Info("start")
	defer logger.Info("end")
	response := InfoResponse{}
	client.Invoker.Command("info")
	err := client.Invoker.Execute(logger, &response)
	if err != nil {
		return InfoResponse{}, err
	}
	return response, nil
}
