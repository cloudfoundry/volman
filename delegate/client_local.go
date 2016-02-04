package delegate

import "github.com/cloudfoundry-incubator/volman"

type LocalClient struct {
}

func NewLocalClient() *LocalClient {
	return &LocalClient{}
}

func (client *LocalClient) ListDrivers() (volman.ListDriversResponse, error) {
	return volman.ListDriversResponse{}, nil
}
