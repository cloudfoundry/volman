package volhttp

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"code.cloudfoundry.org/cfhttp"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/volman"
	"github.com/tedsuo/rata"
)

type remoteClient struct {
	HttpClient *http.Client
	reqGen     *rata.RequestGenerator
}

func NewRemoteClient(volmanURL string) *remoteClient {
	return &remoteClient{
		HttpClient: cfhttp.NewClient(),
		reqGen:     rata.NewRequestGenerator(volmanURL, volman.Routes),
	}
}

func (r *remoteClient) ListDrivers(logger lager.Logger) (volman.ListDriversResponse, error) {
	logger = logger.Session("list-drivers")
	logger.Debug("start")
	defer logger.Debug("end")

	request, err := r.reqGen.CreateRequest(volman.ListDriversRoute, nil, nil)
	if err != nil {
		return volman.ListDriversResponse{}, r.clientError(logger, err, "error-creating-request", lager.Data{"route": volman.ListDriversRoute})
	}

	response, err := r.HttpClient.Do(request)
	if err != nil {
		return volman.ListDriversResponse{}, r.clientError(logger, err, "error-listing-drivers", lager.Data{})
	}
	var drivers volman.ListDriversResponse
	err = unmarshallJSON(logger, response.Body, &drivers)

	if err != nil {
		return volman.ListDriversResponse{}, r.clientError(logger, err, "error-parsing-json", lager.Data{})
	}

	return drivers, err
}

func (r *remoteClient) Mount(logger lager.Logger, driverId string, volumeId string, config map[string]interface{}) (volman.MountResponse, error) {
	logger = logger.Session("mount")
	logger.Info("start")
	defer logger.Info("end")

	MountRequest := volman.MountRequest{driverId, volumeId, config}

	sendingJson, err := json.Marshal(MountRequest)
	if err != nil {
		return volman.MountResponse{}, r.clientError(logger, err, "error-marshalling-json", lager.Data{"mount-request": MountRequest})
	}

	request, err := r.reqGen.CreateRequest(volman.MountRoute, nil, bytes.NewReader(sendingJson))

	if err != nil {
		return volman.MountResponse{}, r.clientError(logger, err, "error-creating-request", lager.Data{"route": volman.MountRoute})
	}

	response, err := r.HttpClient.Do(request)
	if err != nil {
		return volman.MountResponse{}, r.clientError(logger, err, "error-mounting-volume", lager.Data{"volume-id": volumeId})
	}

	if response.StatusCode == 500 {
		var remoteError volman.Error
		if err := unmarshallJSON(logger, response.Body, &remoteError); err != nil {
			return volman.MountResponse{}, r.clientError(logger, err, "error-parsing-response", lager.Data{"route": volman.MountRoute})
		}
		return volman.MountResponse{}, remoteError
	}

	var mountPoint volman.MountResponse
	if err := unmarshallJSON(logger, response.Body, &mountPoint); err != nil {
		return volman.MountResponse{}, r.clientError(logger, err, "error-parsing-response", lager.Data{"route": volman.MountRoute})
	}

	return mountPoint, err
}

func (r *remoteClient) Unmount(logger lager.Logger, driverId string, volumeId string) error {
	logger = logger.Session("mount")
	logger.Info("start")
	defer logger.Info("end")

	unmountRequest := volman.UnmountRequest{driverId, volumeId}
	payload, err := json.Marshal(unmountRequest)
	if err != nil {
		return r.clientError(logger, err, "error-marshalling-json", lager.Data{"unmount-request": unmountRequest})
	}

	request, err := r.reqGen.CreateRequest(volman.UnmountRoute, nil, bytes.NewReader(payload))

	if err != nil {
		return r.clientError(logger, err, "error-creating-request", lager.Data{"route": volman.UnmountRoute})
	}

	response, err := r.HttpClient.Do(request)
	if err != nil {
		return r.clientError(logger, err, "error-unmounting-volume", lager.Data{"volume-id": volumeId})
	}

	if response.StatusCode == 500 {
		var remoteError volman.Error
		if err := unmarshallJSON(logger, response.Body, &remoteError); err != nil {
			return r.clientError(logger, err, "error-parsing-response", lager.Data{"route": volman.UnmountRoute})
		}
		return remoteError
	}

	return nil
}

func unmarshallJSON(logger lager.Logger, reader io.ReadCloser, jsonResponse interface{}) error {
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		logger.Error("error-reading-http-response-body", err)
	}
	err = json.Unmarshal(body, jsonResponse)

	return err
}

func (r *remoteClient) clientError(logger lager.Logger, err error, msg string, data lager.Data) error {
	logger.Error(msg, err, data)
	return err
}
