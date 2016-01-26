package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	cf_debug_server "github.com/cloudfoundry-incubator/cf-debug-server"
	cf_lager "github.com/cloudfoundry-incubator/cf-lager"
	"github.com/pivotal-golang/lager"
	. "github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
	"github.com/tedsuo/rata"
)

var atAddress = flag.String(
	"listenAddr",
	"0.0.0.0:8750",
	"host:port to serve volume management functions",
)

func main() {
	withLogger, logTap := logger()
	cf_debug_server.AddFlags(flag.CommandLine)
	flag.Parse()

	servers := grouper.Members{}
	servers = append(server(*atAddress), servers...)

	atDebugAddress := cf_debug_server.DebugAddress(flag.CommandLine)

	if atDebugAddress != "" {
		servers = append(debugServer(atDebugAddress, logTap), servers...)
	}

	processes := sigmon.New(grouper.NewOrdered(os.Interrupt, servers))

	run(Invoke(processes), withLogger)
}

func run(process Process, logger lager.Logger) {
	logger.Info("started")
	err := <-process.Wait()
	if err != nil {
		os.Exit(1)
	}
	logger.Info("exited")
}

func debugServer(atDebugAddress string, logTap *lager.ReconfigurableSink) grouper.Members {
	return grouper.Members{
		{"debug-server", cf_debug_server.Runner(atDebugAddress, logTap)},
	}
}

func server(atAddress string) grouper.Members {
	handler, _ := volmanHandlers()
	server := http_server.New(atAddress, handler)
	return grouper.Members{
		{"volman-server", server},
	}
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

func logger() (lager.Logger, *lager.ReconfigurableSink) {
	cf_lager.AddFlags(flag.CommandLine)
	logger, reconfigurableSink := cf_lager.New("volman")
	return logger, reconfigurableSink
}
