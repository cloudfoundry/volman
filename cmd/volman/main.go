package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/rata"
)

var listenAddr = flag.String(
	"listenAddr",
	"0.0.0.0:8750",
	"host:port to serve volume management functions",
)

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

func main() {
	flag.Parse()
	handler, _ := volmanHandlers()
	volmanServer := http_server.New(*listenAddr, handler)

	process := ifrit.Invoke(volmanServer)

	fmt.Println("volman.started")

	err := <-process.Wait()
	if err != nil {
		os.Exit(1)
	}

	fmt.Println("volman.exited")
}
