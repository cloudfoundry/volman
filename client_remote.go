package volman

import (
	"github.com/cloudfoundry-incubator/cf_http"
	. "github.com/cloudfoundry-incubator/volman/remote"
)

type RemoteClient struct {
	RemoteHttpClient_ RemoteHttpClientInterface
}

func NewRemoteClient(volmanURL string) *RemoteClient {
	return &RemoteClient{
		RemoteHttpClient_: &RemoteHttpClient{
			HttpClient_: cf_http.NewClient(),
			URL_:        volmanURL,
		},
	}
}

func (client *RemoteClient) ListDrivers() (Drivers, error) {
	var drivers Drivers
	request := "/v1/drivers"
	err := client.RemoteHttpClient_.Get(request).AndReturnJsonIn(&drivers)
	return drivers, err
}
