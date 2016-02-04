package volman

import (
	"github.com/cloudfoundry-incubator/volman/remote"
)

type RemoteClient struct {
	RemoteHttpClient remote.RemoteHttpClientInterface
}

func NewRemoteClient(volmanURL string) *RemoteClient {
	return &RemoteClient{
		RemoteHttpClient: remote.NewClient(volmanURL),
	}
}

func (client *RemoteClient) ListDrivers() (ListDriversResponse, error) {
	var drivers ListDriversResponse
	request := "/v1/drivers"
	err := client.RemoteHttpClient.Get(request).AndReturnJsonIn(&drivers)
	return drivers, err
}
