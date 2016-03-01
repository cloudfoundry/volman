package voldriver

import(
	"github.com/pivotal-golang/lager"
)

type InfoResponse struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

//go:generate counterfeiter -o ../volmanfakes/fake_driver_client.go . Driver

type Driver interface {
	Info(logger lager.Logger) (InfoResponse, error)
	Mount(logger lager.Logger, mountRequest MountRequest) (MountResponse, error)
}

type MountRequest struct {
	VolumeId string `json:"volumeId"`
	Config   string `json:"config"`
}

type MountResponse struct {
	Path string `json:"path"`
}