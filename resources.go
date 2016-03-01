package volman

import (
	"github.com/tedsuo/rata"
	"github.com/cloudfoundry-incubator/volman/voldriver"
)


const (
	ListDriversRoute = "drivers"
	MountRoute       = "mount"
)

var Routes = rata.Routes{
	{Path: "/drivers", Method: "GET", Name: ListDriversRoute},
	{Path: "/drivers/mount", Method: "POST", Name: MountRoute},
}



type ListDriversResponse struct {
	Drivers []voldriver.InfoResponse `json:"drivers"`
}

type MountRequest struct {
	DriverId string `json:"driverId"`
	VolumeId string `json:"volumeId"`
	Config   string `json:"config"`
}

type MountResponse struct {
	Path string `json:"path"`
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
