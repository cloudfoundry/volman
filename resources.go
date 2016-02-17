package volman

import "github.com/tedsuo/rata"

const (
	ListDriversRoute = "drivers"
)

var Routes = rata.Routes{
	{Path: "/v1/drivers", Method: "GET", Name: ListDriversRoute},
}

type Driver struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ListDriversResponse struct {
	Drivers []Driver `json:"drivers"`
}
