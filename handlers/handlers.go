package volman

import (
	"net/http"

	"github.com/cloudfoundry-incubator/cf_http"
	. "github.com/cloudfoundry-incubator/volman/delegate"
	"github.com/tedsuo/rata"
)

func Generate() (http.Handler, error) {
	var routes = rata.Routes{
		{Path: "/v1/drivers", Method: "GET", Name: "drivers"},
	}

	var handlers = rata.Handlers{
		"drivers": http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			client := NewLocalClient()
			drivers, _ := client.ListDrivers()
			cf_http.WriteJSONResponse(w, http.StatusOK, drivers)
		}),
	}

	return rata.NewRouter(routes, handlers)
}
