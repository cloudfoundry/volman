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
	logger.Session("list-drivers")
	logger.Info("start")

	request, err := r.reqGen.CreateRequest(ListDriversRoute, nil, nil)

	response, err := r.Get(logger, request)
	if err != nil {
		logger.Fatal("Error in Listing Drivers", err)
	}
	var drivers ListDriversResponse
	err = umarshallJSON(logger, response.Body, &drivers)

	if err != nil {
		logger.Fatal("Error in Parsing JSON Response of List Drivers", err)
	}
	logger.Info("complete")
	return drivers, err
}

func (r *remoteClient) request(logger lager.Logger, operation operationType, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(operation.Method, r.URL+path, body)
	if err != nil {
		logger.Fatal("Error in creating HTTP Request", err)
	}
	for header, value := range operation.Headers {
		req.Header.Add(header, value)
	}

	response, err := r.HttpClient.Do(req)
	return response, err

}

func (r *remoteClient) Get(logger lager.Logger, fromPath string) (*http.Response, error) {
	return r.request(logger, usingGet, fromPath, nil)
}

func umarshallJSON(logger lager.Logger, reader io.ReadCloser, jsonResponse interface{}) error {

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		logger.Fatal("Error in Reading HTTP Response body", err)
	}
	err = json.Unmarshal(body, jsonResponse)

	return err
}
