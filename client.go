package volman

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	Mount(logger lager.Logger, driver Driver, volumeId string, config string) (MountPointResponse, error)
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
	if err != nil {
		return ListDriversResponse{}, r.clientError(logger, err, fmt.Sprintf("Error creating request to %s", ListDriversRoute))
	}

	response, err := r.HttpClient.Do(request)
	if err != nil {
		return ListDriversResponse{}, r.clientError(logger, err, "Error in Listing Drivers remote call")
	}
	var drivers ListDriversResponse
	err = unmarshallJSON(logger, response.Body, &drivers)

	if err != nil {
		return ListDriversResponse{}, r.clientError(logger, err, "Error in Parsing JSON Response of List Drivers")
	}
	logger.Info("complete")
	return drivers, err
}

func (r *remoteClient) Mount(logger lager.Logger, driver Driver, volumeId string, config string) (MountPointResponse, error) {
	logger = logger.Session("mount")
	logger.Info("start")
	defer logger.Info("complete")

	mountPointRequest := MountPointRequest{driver, volumeId, config}

	sendingJson, err := json.Marshal(mountPointRequest)
	if err != nil {
		return MountPointResponse{}, r.clientError(logger, err, fmt.Sprintf("Error marshalling JSON request %#v", mountPointRequest))
	}

	request, err := r.reqGen.CreateRequest(MountRoute, nil, bytes.NewReader(sendingJson))
	if err != nil {
		return MountPointResponse{}, r.clientError(logger, err, fmt.Sprintf("Error creating request to %s", MountRoute))
	}

	response, err := r.HttpClient.Do(request)
	if err != nil {
		return MountPointResponse{}, r.clientError(logger, err, fmt.Sprintf("Error mounting volume %s", volumeId))
	}

	if response.StatusCode == 500 {
		var remoteError Error
		if err := unmarshallJSON(logger, response.Body, &remoteError); err != nil {
			return MountPointResponse{}, r.clientError(logger, err, fmt.Sprintf("Error parsing 500 response from %s", MountRoute))
		}
		return MountPointResponse{}, remoteError
	}

	var mountPoint MountPointResponse
	if err := unmarshallJSON(logger, response.Body, &mountPoint); err != nil {
		return MountPointResponse{}, r.clientError(logger, err, fmt.Sprintf("Error parsing response from %s", MountRoute))
	}

	return mountPoint, err
}

func unmarshallJSON(logger lager.Logger, reader io.ReadCloser, jsonResponse interface{}) error {
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		logger.Error("Error in Reading HTTP Response body from remote.", err)
	}
	err = json.Unmarshal(body, jsonResponse)

	return err
}

func (r *remoteClient) clientError(logger lager.Logger, err error, msg string) error {
	logger.Error(msg, err)
	return err
}
