package volhttp

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	cf_http_handlers "code.cloudfoundry.org/cfhttp/handlers"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/volman"
	"github.com/tedsuo/rata"
)

func respondWithError(logger lager.Logger, info string, err error, w http.ResponseWriter, data lager.Data) {
	logger.Error(info, err, data)
	cf_http_handlers.WriteJSONResponse(w, http.StatusInternalServerError, volman.NewError(err))
}

func NewHandler(logger lager.Logger, client volman.Manager) (http.Handler, error) {
	logger = logger.Session("server")
	logger.Info("start")
	defer logger.Info("end")

	var handlers = rata.Handlers{
		"drivers": newListDriversHandler(logger, client),
		"mount":   newMountHandler(logger, client),
		"unmount": newUnmountHandler(logger, client),
	}

	return rata.NewRouter(volman.Routes, handlers)
}

func newListDriversHandler(logger lager.Logger, client volman.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		logger := logger.Session("drivers")
		logger.Info("start")
		defer logger.Info("end")

		drivers, _ := client.ListDrivers(logger) // <- fix this!
		cf_http_handlers.WriteJSONResponse(w, http.StatusOK, drivers)
	}
}

func newMountHandler(logger lager.Logger, client volman.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		logger := logger.Session("mount")
		logger.Info("start")
		defer logger.Info("end")

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			respondWithError(logger, "error-reading-mount-request", err, w, lager.Data{})
			return
		}

		var mountRequest volman.MountRequest
		if err = json.Unmarshal(body, &mountRequest); err != nil {
			respondWithError(logger, "error-reading-mount-request", err, w, lager.Data{"body": body})
			return
		}

		mountPoint, err := client.Mount(logger, mountRequest.DriverId, mountRequest.VolumeId, mountRequest.Config)
		if err != nil {
			respondWithError(logger, "error-mounting-volume", err, w, lager.Data{"volumeId": mountRequest.VolumeId, "driverId": mountRequest.DriverId})
			return
		}

		cf_http_handlers.WriteJSONResponse(w, http.StatusOK, mountPoint)
	}
}

func newUnmountHandler(logger lager.Logger, client volman.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		logger := logger.Session("unmount")
		logger.Info("start")
		defer logger.Info("end")

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			respondWithError(logger, "error-reading-unmount-request", err, w, lager.Data{})
			return
		}

		var unmountRequest volman.UnmountRequest
		if err = json.Unmarshal(body, &unmountRequest); err != nil {
			respondWithError(logger, "error-reading-unmount-request", err, w, lager.Data{"body": body})
			return
		}

		err = client.Unmount(logger, unmountRequest.DriverId, unmountRequest.VolumeId)
		if err != nil {
			respondWithError(logger, "error-unmounting-volume", err, w, lager.Data{"volumeId": unmountRequest.VolumeId, "driverId": unmountRequest.DriverId})
			return
		}

		cf_http_handlers.WriteJSONResponse(w, http.StatusOK, struct{}{})
	}
}
