package handlers

import (
	"net/http"

	cf_http_handlers "github.com/cloudfoundry-incubator/cf_http/handlers"
	"github.com/cloudfoundry-incubator/volman/delegate"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/rata"
)

func New(logger lager.Logger) (http.Handler, error) {

	var handlers = rata.Handlers{
		"drivers": http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			client := delegate.NewLocalClient()
			drivers, _ := client.ListDrivers(logger)
			cf_http_handlers.WriteJSONResponse(w, http.StatusOK, drivers)
		}),
	}

	return rata.NewRouter(cf_http_handlers.Routes, handlers)
}
