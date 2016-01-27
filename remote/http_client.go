package remote

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/cloudfoundry-incubator/cf_http"
	"github.com/errors"
)

type operationType struct {
	Method  string
	Headers map[string]string
}

var (
	usingGet  = operationType{"GET", map[string]string{"Accept": "application/json"}}
	usingPost = operationType{"POST", map[string]string{"Accept": "application/json", "Content-Type": "application/json"}}
)

type RemoteHttpClient struct {
	HttpClient_ HttpClient
	URL_        string
}

func NewClient(url string) RemoteHttpClientInterface {
	return &RemoteHttpClient{
		HttpClient_: cf_http.NewClient(),
		URL_:        url,
	}
}

func (withAPI *RemoteHttpClient) URL() string { return withAPI.URL_ }

func (withAPI *RemoteHttpClient) request(operation operationType, path string, body io.Reader) RequestReturnInterface {
	req, err := http.NewRequest(operation.Method, withAPI.URL()+path, body)
	if err != nil {
		err = errors.Wrap(err, 1)
		return &RequestReturn{nil, err}
	}
	for header, value := range operation.Headers {
		req.Header.Add(header, value)
	}
	var toReturn RequestReturn
	if withAPI.HttpClient_ == nil {
		withAPI.HttpClient_ = http.DefaultClient
	}
	toReturn.Response, toReturn.Err = withAPI.HttpClient_.Do(req)
	return &toReturn
}

func (withAPI *RemoteHttpClient) Get(fromPath string) RequestReturnInterface {
	return withAPI.request(usingGet, fromPath, nil)
}

type RequestReturn struct {
	Response *http.Response
	Err      error
}

func (request *RequestReturn) AndReturnJsonIn(jsonResponse interface{}) (err error) {
	if request.Err != nil {
		return errors.Wrap(request.Err, 1)
	}

	body, err := ioutil.ReadAll(request.Response.Body)
	if err != nil {
		return errors.Wrap(err, 1)
		return
	}
	err = json.Unmarshal(body, jsonResponse)
	if err != nil {
		return errors.Wrap(fmt.Errorf("Unmarshalling JSON (%s): %s", err.Error(), body), 1)
	}
	return
}
