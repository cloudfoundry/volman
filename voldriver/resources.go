package voldriver

import (
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/rata"
)

const (
	CreateRoute   = "create"
	MountRoute    = "mount"
	UnmountRoute  = "unmount"
	RemoveRoute   = "remove"
	GetRoute      = "get"
	ActivateRoute = "activate"
)

var Routes = rata.Routes{
	{Path: "/VolumeDriver.Create", Method: "POST", Name: CreateRoute},
	{Path: "/VolumeDriver.Mount", Method: "POST", Name: MountRoute},
	{Path: "/VolumeDriver.Unmount", Method: "POST", Name: UnmountRoute},
	{Path: "/VolumeDriver.Remove", Method: "POST", Name: RemoveRoute},
	{Path: "/VolumeDriver.Get", Method: "GET", Name: GetRoute},
	{Path: "/Plugin.Activate", Method: "POST", Name: ActivateRoute},
}

//go:generate counterfeiter -o ../volmanfakes/fake_driver_client.go . Driver

type Driver interface {
	Activate(logger lager.Logger) ActivateResponse
	Create(logger lager.Logger, createRequest CreateRequest) ErrorResponse
	Mount(logger lager.Logger, mountRequest MountRequest) MountResponse
	Unmount(logger lager.Logger, unmountRequest UnmountRequest) ErrorResponse
	Remove(logger lager.Logger, removeRequest RemoveRequest) ErrorResponse
	Get(logger lager.Logger, getRequest GetRequest) GetResponse
}

type ActivateResponse struct {
	Err        string
	Implements []string
}

type CreateRequest struct {
	Name string
	Opts map[string]interface{}
}

type MountRequest struct {
	Name string
}

type MountResponse struct {
	Err        string
	Mountpoint string
}

type UnmountRequest struct {
	Name string
}

type RemoveRequest struct {
	Name string
}

type ErrorResponse struct {
	Err string
}

type GetRequest struct {
	Name string
}

type GetResponse struct {
	Volume VolumeInfo
	Err    string
}

type VolumeInfo struct {
	Name       string
	Mountpoint string
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
