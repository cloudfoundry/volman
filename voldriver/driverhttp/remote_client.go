package driverhttp

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"strings"

	"fmt"

	"code.cloudfoundry.org/cfhttp"
	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/rata"

	os_http "net/http"

	"time"

	"errors"

	"github.com/cloudfoundry/gunk/http_wrap"
	"github.com/pivotal-golang/clock"
)

type reqFactory struct {
	reqGen  *rata.RequestGenerator
	route   string
	payload []byte
}

func newReqFactory(reqGen *rata.RequestGenerator, route string, payload []byte) *reqFactory {
	return &reqFactory{
		reqGen:  reqGen,
		route:   route,
		payload: payload,
	}
}

func (r *reqFactory) Request() (*os_http.Request, error) {
	return r.reqGen.CreateRequest(r.route, nil, bytes.NewBuffer(r.payload))
}

type remoteClient struct {
	HttpClient http_wrap.Client
	reqGen     *rata.RequestGenerator
	clock      clock.Clock
}

func NewRemoteClient(url string, tls *voldriver.TLSConfig) (*remoteClient, error) {
	baseClient := cfhttp.NewClient()
	httpClient := http_wrap.NewClientFrom(baseClient)

	if strings.Contains(url, ".sock") {
		httpClient = cfhttp.NewUnixClient(url)
		url = fmt.Sprintf("unix://%s", url)
	} else {
		if tls != nil {
			tlsConfig, err := cfhttp.NewTLSConfig(tls.CertFile, tls.KeyFile, tls.CAFile)
			if err != nil {
				return nil, err
			}

			tlsConfig.InsecureSkipVerify = tls.InsecureSkipVerify

			if tr, ok := baseClient.Transport.(*http.Transport); ok {
				tr.TLSClientConfig = tlsConfig
			} else {
				return nil, errors.New("Invalid transport")
			}
		}

	}

	return NewRemoteClientWithClient(url, httpClient, clock.NewClock()), nil
}
func NewRemoteClientWithClient(socketPath string, client http_wrap.Client, clock clock.Clock) *remoteClient {
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

	request := newReqFactory(r.reqGen, voldriver.ActivateRoute, nil)

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
		logger.Error("failed-parsing-activate-response", err)
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

	request := newReqFactory(r.reqGen, voldriver.CreateRoute, payload)

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

func (r *remoteClient) List(logger lager.Logger) voldriver.ListResponse {
	logger = logger.Session("remoteclient-list")
	logger.Info("start")
	defer logger.Info("end")

	request := newReqFactory(r.reqGen, voldriver.ListRoute, nil)

	response, err := r.do(logger, request)
	if err != nil {
		logger.Error("failed-list", err)
		return voldriver.ListResponse{Err: err.Error()}
	}

	if response.StatusCode == 500 {
		var remoteError voldriver.ListResponse
		if err := unmarshallJSON(logger, response.Body, &remoteError); err != nil {
			logger.Error("failed-parsing-error-response", err)
			return voldriver.ListResponse{Err: err.Error()}
		}
		return remoteError
	}

	var list voldriver.ListResponse
	if err := unmarshallJSON(logger, response.Body, &list); err != nil {
		logger.Error("failed-parsing-list-response", err)
		return voldriver.ListResponse{Err: err.Error()}
	}

	return list
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

	request := newReqFactory(r.reqGen, voldriver.MountRoute, sendingJson)

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

func (r *remoteClient) Path(logger lager.Logger, pathRequest voldriver.PathRequest) voldriver.PathResponse {
	logger = logger.Session("path")
	logger.Info("start")
	defer logger.Info("end")

	payload, err := json.Marshal(pathRequest)
	if err != nil {
		logger.Error("failed-marshalling-request", err)
		return voldriver.PathResponse{Err: err.Error()}
	}

	request := newReqFactory(r.reqGen, voldriver.PathRoute, payload)

	response, err := r.do(logger, request)
	if err != nil {
		logger.Error("failed-volume-path", err)
		return voldriver.PathResponse{Err: err.Error()}
	}

	if response.StatusCode == 500 {
		var remoteErrorResponse voldriver.PathResponse
		if err := unmarshallJSON(logger, response.Body, &remoteErrorResponse); err != nil {
			logger.Error("failed-parsing-error-response", err)
			return voldriver.PathResponse{Err: err.Error()}
		}
		return remoteErrorResponse
	}

	var mountPoint voldriver.PathResponse
	if err := unmarshallJSON(logger, response.Body, &mountPoint); err != nil {
		logger.Error("failed-parsing-path-response", err)
		return voldriver.PathResponse{Err: err.Error()}
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

	request := newReqFactory(r.reqGen, voldriver.UnmountRoute, payload)

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

	request := newReqFactory(r.reqGen, voldriver.RemoveRoute, payload)

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

	request := newReqFactory(r.reqGen, voldriver.GetRoute, payload)

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

func (r *remoteClient) do(logger lager.Logger, requestFactory *reqFactory) (*os_http.Response, error) {
	var response *os_http.Response

	backoff := newExponentialBackOff(30*time.Second, r.clock)

	err := backoff.Retry(func() error {
		var (
			err     error
			request *os_http.Request
		)

		request, err = requestFactory.Request()
		if err != nil {
			logger.Error("request-gen-failed", err)
			return err
		}

		response, err = r.HttpClient.Do(request)
		if err != nil {
			logger.Error("request-failed", err)
			return err
		}
		logger.Debug("response", lager.Data{"response": response.Status})

		return err
	})

	return response, err
}
