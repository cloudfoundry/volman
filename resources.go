package volman

import "github.com/tedsuo/rata"

const (
	ListDriversRoute = "drivers"
)

var Routes = rata.Routes{
	{Path: "/v1/drivers", Method: "GET", Name: ListDriversRoute},
}

type ListDriversResponse struct {
	Drivers []string `json:"drivers"`
}
