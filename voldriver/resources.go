package voldriver

import (
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/rata"
)

const (
	MountRoute   = "mount"
	UnmountRoute = "unmount"
)

var Routes = rata.Routes{
	{Path: "/mount", Method: "POST", Name: MountRoute},
	{Path: "/unmount", Method: "POST", Name: UnmountRoute},
}

//go:generate counterfeiter -o ../volmanfakes/fake_driver_client.go . Driver

type Driver interface {
	Info(logger lager.Logger) (InfoResponse, error)
	//	Create(logger lager.Logger, createRequest) error
	Mount(logger lager.Logger, mountRequest MountRequest) (MountResponse, error)
	Unmount(logger lager.Logger, unmountRequest UnmountRequest) error
	//Remove(logger lager.Logger, removeRequest) error
}

type InfoResponse struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

type MountRequest struct {
	VolumeId string `json:"volumeId"`
	Config   string `json:"config"`
}

type MountResponse struct {
	Path string `json:"path"`
}

type UnmountRequest struct {
	VolumeId string `json:"volumeId"`
}

func NewError(err error) Error {
	return Error{err.Error()}
}

type Error struct {
	Description string `json:"description"`
}

func (e Error) Error() string {
	return e.Description
}

type DriverSpec struct {
	Name    string `json:"Name"`
	Address string `json:"Addr"`
}
