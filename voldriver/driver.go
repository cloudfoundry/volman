package voldriver
import (
	"github.com/cloudfoundry-incubator/volman/system"
	"github.com/cloudfoundry-incubator/volman"
	"github.com/pivotal-golang/lager"
	"fmt"
)

type DriverClientCli struct {
	UseExec     system.Exec
	DriversPath string
	Name string
}

func NewDriverClientCli(path string, useExec system.Exec, name string) *DriverClientCli{
	return &DriverClientCli{
		UseExec:     useExec,
		DriversPath: path,
		Name: name,
	}
}

func (client *DriverClientCli) Mount(logger lager.Logger, volumeId string, config string) (string, error) {
	logger.Info("start")
	defer logger.Info("end")
	response := struct {
		Path string `json:"path"`
	}{}
	invoker := NewCliInvoker(client.UseExec, fmt.Sprintf("%s/%s",client.DriversPath,client.Name), "mount", volumeId, config)
	err := invoker.InvokeDriver(logger, &response)
	if err != nil {
		return "", err
	}

	return response.Path, nil
}

	
func (client *DriverClientCli) Info(logger lager.Logger) (volman.DriverInfo, error) {
	logger.Info("start")
	defer logger.Info("end")
	response := volman.DriverInfo{}
	invoker := NewCliInvoker(client.UseExec, fmt.Sprintf("%s/%s",client.DriversPath,client.Name), "info")
	err := invoker.InvokeDriver(logger, &response)
	if err != nil {
		return volman.DriverInfo{}, err
	}
	return response, nil
}