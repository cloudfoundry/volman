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

	os_http "net/http"

	"time"

	"github.com/pivotal-golang/clock"
)

type remoteClient struct {
	HttpClient http.Client
	reqGen     *rata.RequestGenerator
	clock      clock.Clock
}

func NewRemoteClient(url string) *remoteClient {
	httpClient := http.NewClientFrom(cf_http.NewClient())

	if strings.Contains(url, ".sock") {
		httpClient = cf_http.NewUnixClient(url)
		url = fmt.Sprintf("unix://%s", url)
	}

	return NewRemoteClientWithClient(url, httpClient, clock.NewClock())
}

func NewRemoteClientWithClient(socketPath string, client http.Client, clock clock.Clock) *remoteClient {
	return &remoteClient{
		HttpClient: client,
		reqGen:     rata.NewRequestGenerator(socketPath, voldriver.Routes),
		clock:      clock,
	}
}

func (r *remoteClient) Activate(logger lager.Logger) voldriver.ActivateResponse {
	logger = logger.Session("activate")
	logger.Info("start")
	defer logger.Info("end")

	request, err := r.reqGen.CreateRequest(voldriver.ActivateRoute, nil, nil)
	if err != nil {
		logger.Error("failed-creating-request", err)
		return voldriver.ActivateResponse{Err: err.Error()}
	}

	response, err := r.do(logger, request)
	if err != nil {
		logger.Error("failed-activate", err)
		return voldriver.ActivateResponse{Err: err.Error()}
	}

	if response.StatusCode == 500 {
		var remoteError voldriver.ActivateResponse
		if err := unmarshallJSON(logger, response.Body, &remoteError); err != nil {
			logger.Error("failed-parsing-error-response", err)
			return voldriver.ActivateResponse{Err: err.Error()}
		}
		return remoteError
	}

	var activate voldriver.ActivateResponse
	if err := unmarshallJSON(logger, response.Body, &activate); err != nil {
		logger.Error("failed-parsing-info-response", err)
		return voldriver.ActivateResponse{Err: err.Error()}
	}

	return activate
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
	response, err := r.do(logger, request)
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

	response, err := r.do(logger, request)
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

	response, err := r.do(logger, request)
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

	response, err := r.do(logger, request)
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

	response, err := r.do(logger, request)
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
func (r *remoteClient) do(logger lager.Logger, request *os_http.Request) (*os_http.Response, error) {
	var response *os_http.Response

	// NewExponentialBackOff creates an instance of ExponentialBackOff using default values.
	customBackoff := NewExponentialBackOff(r.clock)
	customBackoff.MaxElapsedTime = 30 * time.Second

	count := 0
	err := Retry(func() error {
		var err error
		logger.Info("Trying to contact", lager.Data{"count": count})
		response, err = r.HttpClient.Do(request)
		count = count + 1

		if response == nil {
			logger.Info("response-nil")
		} else {

			logger.Info("response", lager.Data{"response": response.Status})
		}
		logger.Error("Retry", err)
		return err
	}, customBackoff)

	return response, err
}
