package main

import (
	"flag"
	"fmt"
	"net/http"
	"sync"
	"time"

	// cf_debug_server "github.com/cloudfoundry-incubator/cf-debug-server"
	// cf_lager "github.com/cloudfoundry-incubator/cf-lager"
	// "github.com/pivotal-golang/lager"
)

// var communicationTimeout = flag.Duration(
// 	"communicationTimeout",
// 	10*time.Second,
// 	"Timeout applied to all HTTP requests.",
// )
var listenAddr = flag.String(
	"listenAddr",
	"0.0.0.0:8750",
	"host:port to serve volume management functions",
)
var dropsondePort = flag.Int(
	"dropsondePort",
	3457,
	"port the local metron agent is listening on",
)

func main() {
	// cf_debug_server.AddFlags(flag.CommandLine)
	// cf_lager.AddFlags(flag.CommandLine)

	flag.Parse()
	//logger, reconfigurableSink := cf_lager.New("volman")
	//initializeDropsonde(logger)

	http.HandleFunc("/v1/drivers", handler)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(time.Millisecond * 10000)
		http.ListenAndServe(fmt.Sprintf("%s", *listenAddr), nil)
	}()
	//cf_http.Initialize(*communicationTimeout)
	fmt.Println("volman.started")
	wg.Wait()
}

// func initializeVolmanServer(runner volmantypes.Runner) ifrit.Runner {
// 	return http_server.New(*listenAddr, handlers.New(runner))
// }

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "none")
}

// func initializeDropsonde(logger lager.Logger) {
// 	dropsondeDestination := fmt.Sprint("localhost:", *dropsondePort)
// 	err := dropsonde.Initialize(dropsondeDestination, dropsondeOrigin)
// 	if err != nil {
// 		logger.Error("failed to initialize dropsonde: %v", err)
// 	}
// }
