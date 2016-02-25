package volman

import "github.com/tedsuo/rata"

const (
	ListDriversRoute = "drivers"
	MountRoute       = "mount"
)

var Routes = rata.Routes{
	{Path: "/drivers", Method: "GET", Name: ListDriversRoute},
	{Path: "/drivers/mount/", Method: "POST", Name: MountRoute},
}

type Driver struct {
	Name   string `json:"name"`
	Binary string `json:"binary,omitempty"`
}

type ListDriversResponse struct {
	Drivers []DriverInfo `json:"drivers"`
}

type DriverInfo struct {
	Name string `json:"name"`
}

type MountPointRequest struct {
	DriverId string `json:"driverId"`
	VolumeId string `json:"volumeId"`
	Config   string `json:"config"`
}

type MountPointResponse struct {
	Path string `json:"path"`
}

func ErrorFrom(err error) Error {
	return Error{err.Error()}
}

type Error struct {
	Description string `json:"description"`
}

func (e Error) Error() string {
	return e.Description
}
