package volhttp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	cf_http_handlers "github.com/cloudfoundry-incubator/cf_http/handlers"
	"github.com/cloudfoundry-incubator/volman"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/rata"
)

func respondWithError(logger lager.Logger, info string, err error, w http.ResponseWriter) {
	logger.Error(info, err)
	cf_http_handlers.WriteJSONResponse(w, http.StatusInternalServerError, volman.NewError(err))
}

func NewHandler(logger lager.Logger, client volman.Manager) (http.Handler, error) {
	logger = logger.Session("server")
	logger.Info("start")
	defer logger.Info("end")
	var handlers = rata.Handlers{
		"drivers": http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			drivers, _ := client.ListDrivers(logger)
			cf_http_handlers.WriteJSONResponse(w, http.StatusOK, drivers)
		}),
		"mount": http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			logger.Info("mount")
			defer logger.Info("mount end")
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				respondWithError(logger, "Error reading mount request body", err, w)
				return
			}

			var MountRequest volman.MountRequest
			if err = json.Unmarshal(body, &MountRequest); err != nil {
				respondWithError(logger, fmt.Sprintf("Error reading mount request body: %#v", body), err, w)
				return
			}

			mountPoint, err := client.Mount(logger, MountRequest.DriverId, MountRequest.VolumeId, MountRequest.Config)
			if err != nil {
				respondWithError(logger, fmt.Sprintf("Error mounting volume %s with driver %s", MountRequest.VolumeId, MountRequest.DriverId), err, w)
				return
			}

			cf_http_handlers.WriteJSONResponse(w, http.StatusOK, mountPoint)
		}),
	}

	return rata.NewRouter(volman.Routes, handlers)
}
