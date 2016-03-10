package driverhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/cloudfoundry-incubator/cf_http"
	"github.com/cloudfoundry-incubator/volman/system/http"
	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/rata"
)

type remoteClient struct {
	HttpClient http.Client
	reqGen     *rata.RequestGenerator
}

func NewRemoteClient(url string) *remoteClient {
	return &remoteClient{
		HttpClient: http.NewClientFrom(cf_http.NewClient()),
		reqGen:     rata.NewRequestGenerator(url, voldriver.Routes),
	}
}

func NewRemoteClientWithHttpClient(url string, httpClient http.Client) *remoteClient {
	return &remoteClient{
		HttpClient: httpClient,
		reqGen:     rata.NewRequestGenerator(url, voldriver.Routes),
	}
}

func (r *remoteClient) Info(logger lager.Logger) (voldriver.InfoResponse, error) {
	logger = logger.Session("info")
	logger.Info("start")
	defer logger.Info("end")
	return voldriver.InfoResponse{}, nil
}

func (r *remoteClient) Mount(logger lager.Logger, mountRequest voldriver.MountRequest) (voldriver.MountResponse, error) {
	logger = logger.Session("remoteclient-mount")
	logger.Info("start")
	defer logger.Info("end")

	sendingJson, err := json.Marshal(mountRequest)
	if err != nil { // error path is untestable due to structure type validation in function parameters
		return voldriver.MountResponse{}, r.clientError(logger, err, fmt.Sprintf("Error marshalling JSON request %#v", mountRequest))
	}

	request, err := r.reqGen.CreateRequest(voldriver.MountRoute, nil, bytes.NewReader(sendingJson))
	if err != nil { // error path is untestable due to structure type validation in function parameters
		return voldriver.MountResponse{}, r.clientError(logger, err, fmt.Sprintf("Error creating request to %s", voldriver.MountRoute))
	}

	response, err := r.HttpClient.Do(request)
	if err != nil {
		return voldriver.MountResponse{}, r.clientError(logger, err, fmt.Sprintf("Error mounting volume %s", mountRequest.VolumeId))
	}

	if response.StatusCode == 500 {
		var remoteError voldriver.Error
		if err := unmarshallJSON(logger, response.Body, &remoteError); err != nil {
			return voldriver.MountResponse{}, r.clientError(logger, err, fmt.Sprintf("Error parsing 500 response from %s", voldriver.MountRoute))
		}
		return voldriver.MountResponse{}, remoteError
	}

	var mountPoint voldriver.MountResponse
	if err := unmarshallJSON(logger, response.Body, &mountPoint); err != nil {
		return voldriver.MountResponse{}, r.clientError(logger, err, fmt.Sprintf("Error parsing response from %s", voldriver.MountRoute))
	}

	return mountPoint, err
}

func (r *remoteClient) Unmount(logger lager.Logger, unmountRequest voldriver.UnmountRequest) error {
	logger = logger.Session("mount")
	logger.Info("start")
	defer logger.Info("end")

	payload, err := json.Marshal(unmountRequest)
	if err != nil { // error path is untestable due to structure type validation in function parameters
		return r.clientError(logger, err, fmt.Sprintf("Error marshalling JSON request %#v", unmountRequest))
	}

	request, err := r.reqGen.CreateRequest(voldriver.UnmountRoute, nil, bytes.NewReader(payload))
	if err != nil { // error path is untestable due to structure type validation in function parameters
		return r.clientError(logger, err, fmt.Sprintf("Error creating request to %s", voldriver.UnmountRoute))
	}

	response, err := r.HttpClient.Do(request)
	if err != nil {
		return r.clientError(logger, err, fmt.Sprintf("Error unmounting volume %s", unmountRequest.VolumeId))
	}

	if response.StatusCode == 500 {
		var remoteError voldriver.Error
		if err := unmarshallJSON(logger, response.Body, &remoteError); err != nil {
			return r.clientError(logger, err, fmt.Sprintf("Error parsing 500 response from %s", voldriver.UnmountRoute))
		}
		return remoteError
	}

	return nil

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
