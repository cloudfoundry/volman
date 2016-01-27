package volman

import . "github.com/cloudfoundry-incubator/volman"

type LocalClient struct {
}

func NewLocalClient() *LocalClient {
	return &LocalClient{}
}

func (client *LocalClient) ListDrivers() (Drivers, error) {
	return Drivers{}, nil
}
