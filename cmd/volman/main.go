package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	cf_lager "github.com/cloudfoundry-incubator/cf-lager"
	"github.com/pivotal-golang/lager"
	. "github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/rata"
)

var atAddress = flag.String(
	"listenAddr",
	"0.0.0.0:8750",
	"host:port to serve volume management functions",
)

func main() {
	flag.Parse()
	run(Invoke(server(*atAddress)), withLogger())
}

func run(process Process, logger lager.Logger) {
	logger.Info("volman.started")
	err := <-process.Wait()
	if err != nil {
		os.Exit(1)
	}
	logger.Info("volman.exited")
}

func server(atAddress string) Runner {
	handler, _ := volmanHandlers()
	return http_server.New(atAddress, handler)
}

func volmanHandlers() (http.Handler, error) {
	var routes = rata.Routes{
		{Path: "/v1/drivers", Method: "GET", Name: "drivers"},
	}

	var handlers = rata.Handlers{
		"drivers": http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			fmt.Fprintf(w, "none")
		}),
	}
	return rata.NewRouter(routes, handlers)
}

func withLogger() lager.Logger {
	cf_lager.AddFlags(flag.CommandLine)
	logger, _ := cf_lager.New("volman")
	return logger
}
