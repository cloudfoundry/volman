package driverhttp

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"

	"strings"

	"fmt"

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
	httpClient := http.NewClientFrom(cf_http.NewClient())

	if strings.Contains(url, ".sock") {
		httpClient = cf_http.NewUnixClient(url)
		url = fmt.Sprintf("unix://%s", url)
	}

	return NewRemoteClientWithClient(url, httpClient)
}

func NewRemoteClientWithClient(socketPath string, client http.Client) *remoteClient {
	return &remoteClient{
		HttpClient: client,
		reqGen:     rata.NewRequestGenerator(socketPath, voldriver.Routes),
	}
}

func (r *remoteClient) Info(logger lager.Logger) (voldriver.InfoResponse, error) {
	logger = logger.Session("info")
	logger.Info("start")
	defer logger.Info("end")
	return voldriver.InfoResponse{}, nil
}

func (r *remoteClient) Create(logger lager.Logger, createRequest voldriver.CreateRequest) voldriver.ErrorResponse {
	logger = logger.Session("create", lager.Data{"create_request": createRequest})
	logger.Info("start")
	defer logger.Info("end")

	payload, err := json.Marshal(createRequest)
	if err != nil {
		logger.Error("failed-marshalling-request", err)
		return voldriver.ErrorResponse{Err: err.Error()}
	}

	request, err := r.reqGen.CreateRequest(voldriver.CreateRoute, nil, bytes.NewReader(payload))
	if err != nil {
		logger.Error("failed-creating-request", err)
		return voldriver.ErrorResponse{Err: err.Error()}
	}
	response, err := r.HttpClient.Do(request)
	if err != nil {
		logger.Error("failed-creating-volume", err)
		return voldriver.ErrorResponse{Err: err.Error()}
	}

	if response.StatusCode == 500 {
		var remoteError voldriver.ErrorResponse
		if err := unmarshallJSON(logger, response.Body, &remoteError); err != nil {
			logger.Error("failed-parsing-error-response", err)
			return voldriver.ErrorResponse{Err: err.Error()}
		}
		return remoteError
	}

	return voldriver.ErrorResponse{}
}

func (r *remoteClient) Mount(logger lager.Logger, mountRequest voldriver.MountRequest) voldriver.MountResponse {
	logger = logger.Session("remoteclient-mount", lager.Data{"mount_request": mountRequest})
	logger.Info("start")
	defer logger.Info("end")

	sendingJson, err := json.Marshal(mountRequest)
	if err != nil {
		logger.Error("failed-marshalling-request", err)
		return voldriver.MountResponse{Err: err.Error()}
	}

	request, err := r.reqGen.CreateRequest(voldriver.MountRoute, nil, bytes.NewReader(sendingJson))

	if err != nil {
		logger.Error("failed-creating-request", err)
		return voldriver.MountResponse{Err: err.Error()}
	}

	response, err := r.HttpClient.Do(request)
	if err != nil {
		logger.Error("failed-mounting-volume", err)
		return voldriver.MountResponse{Err: err.Error()}
	}

	if response.StatusCode == 500 {
		var remoteError voldriver.MountResponse
		if err := unmarshallJSON(logger, response.Body, &remoteError); err != nil {
			logger.Error("failed-parsing-error-response", err)
			return voldriver.MountResponse{Err: err.Error()}
		}
		return remoteError
	}

	var mountPoint voldriver.MountResponse
	if err := unmarshallJSON(logger, response.Body, &mountPoint); err != nil {
		logger.Error("failed-parsing-mount-response", err)
		return voldriver.MountResponse{Err: err.Error()}
	}

	return mountPoint
}

func (r *remoteClient) Unmount(logger lager.Logger, unmountRequest voldriver.UnmountRequest) voldriver.ErrorResponse {
	logger = logger.Session("mount")
	logger.Info("start")
	defer logger.Info("end")

	payload, err := json.Marshal(unmountRequest)
	if err != nil {
		logger.Error("failed-marshalling-request", err)
		return voldriver.ErrorResponse{Err: err.Error()}
	}

	request, err := r.reqGen.CreateRequest(voldriver.UnmountRoute, nil, bytes.NewReader(payload))
	if err != nil {
		logger.Error("failed-creating-request", err)
		return voldriver.ErrorResponse{Err: err.Error()}
	}

	response, err := r.HttpClient.Do(request)
	if err != nil {
		logger.Error("failed-unmounting-volume", err)
		return voldriver.ErrorResponse{Err: err.Error()}
	}

	if response.StatusCode == 500 {
		var remoteErrorResponse voldriver.ErrorResponse
		if err := unmarshallJSON(logger, response.Body, &remoteErrorResponse); err != nil {
			logger.Error("failed-parsing-error-response", err)
			return voldriver.ErrorResponse{Err: err.Error()}
		}
		return remoteErrorResponse
	}

	return voldriver.ErrorResponse{}
}

func (r *remoteClient) Remove(logger lager.Logger, removeRequest voldriver.RemoveRequest) voldriver.ErrorResponse {
	logger = logger.Session("remove")
	logger.Info("start")
	defer logger.Info("end")

	payload, err := json.Marshal(removeRequest)
	if err != nil {
		logger.Error("failed-marshalling-request", err)
		return voldriver.ErrorResponse{Err: err.Error()}
	}

	request, err := r.reqGen.CreateRequest(voldriver.RemoveRoute, nil, bytes.NewReader(payload))
	if err != nil {
		logger.Error("failed-creating-request", err)
		return voldriver.ErrorResponse{Err: err.Error()}
	}

	response, err := r.HttpClient.Do(request)
	if err != nil {
		logger.Error("failed-removing-volume", err)
		return voldriver.ErrorResponse{Err: err.Error()}
	}

	if response.StatusCode == 500 {
		var remoteErrorResponse voldriver.ErrorResponse
		if err := unmarshallJSON(logger, response.Body, &remoteErrorResponse); err != nil {
			logger.Error("failed-parsing-error-response", err)
			return voldriver.ErrorResponse{Err: err.Error()}
		}
		return remoteErrorResponse
	}

	return voldriver.ErrorResponse{}
}

func (r *remoteClient) Get(logger lager.Logger, getRequest voldriver.GetRequest) voldriver.GetResponse {
	logger = logger.Session("get")
	logger.Info("start")
	defer logger.Info("end")

	payload, err := json.Marshal(getRequest)
	if err != nil {
		logger.Error("failed-marshalling-request", err)
		return voldriver.GetResponse{Err: err.Error()}
	}

	request, err := r.reqGen.CreateRequest(voldriver.GetRoute, nil, bytes.NewReader(payload))
	if err != nil {
		logger.Error("failed-creating-request", err)
		return voldriver.GetResponse{Err: err.Error()}
	}

	response, err := r.HttpClient.Do(request)
	if err != nil {
		logger.Error("failed-getting-volume", err)
		return voldriver.GetResponse{Err: err.Error()}
	}

	if response.StatusCode == 500 {
		var remoteErrorResponse voldriver.GetResponse
		if err := unmarshallJSON(logger, response.Body, &remoteErrorResponse); err != nil {
			logger.Error("failed-parsing-error-response", err)
			return voldriver.GetResponse{Err: err.Error()}
		}
		return remoteErrorResponse
	}

	return voldriver.GetResponse{}
}

func unmarshallJSON(logger lager.Logger, reader io.ReadCloser, jsonResponse interface{}) error {
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		logger.Error("Error in Reading HTTP Response body from remote.", err)
	}
	err = json.Unmarshal(body, jsonResponse)

	return err
}

func (r *remoteClient) clientError(logger lager.Logger, err error, msg string) string {
	logger.Error(msg, err)
	return err.Error()
}
