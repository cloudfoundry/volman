package volman

type ListDriversResponse struct {
	Drivers []string `json:"drivers"`
}

//go:generate counterfeiter -o volmanfakes/fake_client.go . Client

type Client interface {
	ListDrivers() (ListDriversResponse, error)
}
