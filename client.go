package volman

type Drivers struct {
	Drivers []string `json:"drivers"`
}

//go:generate counterfeiter -o volmanfakes/fake_client.go . Client

type Client interface {
	ListDrivers() (Drivers, error)
}
