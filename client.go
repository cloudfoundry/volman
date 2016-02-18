package volman

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/cloudfoundry-incubator/cf_http"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/rata"
)

type operationType struct {
	Method  string
	Headers map[string]string
}

type Client interface {
	ListDrivers(logger lager.Logger) (ListDriversResponse, error)
}

type remoteClient struct {
	HttpClient *http.Client
	reqGen     *rata.RequestGenerator
}

func NewRemoteClient(volmanURL string) Client {
	return &remoteClient{
		HttpClient: cf_http.NewClient(),
		reqGen:     rata.NewRequestGenerator(volmanURL, Routes),
	}
}

func (r *remoteClient) ListDrivers(logger lager.Logger) (ListDriversResponse, error) {
	logger = logger.Session("list-drivers")
	logger.Info("start")

	request, err := r.reqGen.CreateRequest(ListDriversRoute, nil, nil)

	response, err := r.Get(logger, request)
	if err != nil {
		logger.Fatal("Error in Listing Drivers", err)
	}
	var drivers ListDriversResponse
	err = unmarshallJSON(logger, response.Body, &drivers)

	if err != nil {
		logger.Fatal("Error in Parsing JSON Response of List Drivers", err)
	}
	logger.Info("complete")
	return drivers, err
}

func (r *remoteClient) request(logger lager.Logger, request *http.Request, body io.Reader) (*http.Response, error) {

	response, err := r.HttpClient.Do(request)
	return response, err

}

func (r *remoteClient) Get(logger lager.Logger, request *http.Request) (*http.Response, error) {
	return r.request(logger, request, nil)
}

func unmarshallJSON(logger lager.Logger, reader io.ReadCloser, jsonResponse interface{}) error {

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		logger.Fatal("Error in Reading HTTP Response body", err)
	}
	err = json.Unmarshal(body, jsonResponse)

	return err
}
