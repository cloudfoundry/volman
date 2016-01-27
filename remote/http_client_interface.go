package remote

import "net/http"

//go:generate counterfeiter -o ../volmanfakes/fake_http_client.go . HttpClient

type HttpClient interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

//go:generate counterfeiter -o ../volmanfakes/fake_request_return.go . RequestReturnInterface

type RequestReturnInterface interface {
	AndReturnJsonIn(jsonResponse interface{}) error
}

//go:generate counterfeiter -o ../volmanfakes/fake_remote_http_client.go . RemoteHttpClientInterface

type RemoteHttpClientInterface interface {
	Get(fromPath string) RequestReturnInterface
	URL() string
}
