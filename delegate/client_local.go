package delegate

import (
	"github.com/cloudfoundry-incubator/volman"
	"github.com/pivotal-golang/lager"
)

type LocalClient struct {
}

func NewLocalClient() *LocalClient {
	return &LocalClient{}
}

func (client *LocalClient) ListDrivers(logger lager.Logger) (volman.ListDriversResponse, error) {
	logger.Info("listing-drivers")
	return volman.ListDriversResponse{}, nil
}
