package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	cf_http_handlers "github.com/cloudfoundry-incubator/cf_http/handlers"
	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/delegate"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/rata"
)

func respondWithError(logger lager.Logger, info string, err error, w http.ResponseWriter) {
	logger.Error(info, err)
	cf_http_handlers.WriteJSONResponse(w, http.StatusInternalServerError, volman.ErrorFrom(err))
}

func New(logger lager.Logger, driversPath string) (http.Handler, error) {
	logger = logger.Session("server")
	var handlers = rata.Handlers{
		"drivers": http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			client := delegate.NewLocalClient(driversPath)
			drivers, _ := client.ListDrivers(logger)
			cf_http_handlers.WriteJSONResponse(w, http.StatusOK, drivers)
		}),
		"mount": http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			logger.Info("mount")
			defer logger.Info("mount end")

			client := delegate.NewLocalClient(driversPath)

			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				respondWithError(logger, "Error reading mount request body", err, w)
				return
			}

			var mountPointRequest volman.MountPointRequest
			if err = json.Unmarshal(body, &mountPointRequest); err != nil {
				respondWithError(logger, fmt.Sprintf("Error reading mount request body: %#v", body), err, w)
				return
			}

			mountPoint, err := client.Mount(logger, mountPointRequest.Driver, mountPointRequest.VolumeId, mountPointRequest.Config)
			if err != nil {
				respondWithError(logger, fmt.Sprintf("Error mounting volume %s with driver %s", mountPointRequest.VolumeId, mountPointRequest.Driver.Name), err, w)
				return
			}

			cf_http_handlers.WriteJSONResponse(w, http.StatusOK, mountPoint)
		}),
	}

	return rata.NewRouter(volman.Routes, handlers)
}
