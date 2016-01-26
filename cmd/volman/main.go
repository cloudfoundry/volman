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

var atAddress = flag.String(
	"listenAddr",
	"0.0.0.0:8750",
	"host:port to serve volume management functions",
)

func main() {
	flag.Parse()
	server := server(*atAddress)
	serverProcess := ifrit.Invoke(server)
	run(serverProcess)
}

func run(process ifrit.Process) {
	fmt.Println("volman.started")
	err := <-process.Wait()
	if err != nil {
		os.Exit(1)
	}
	fmt.Println("volman.exited")
}

func server(atAddress string) ifrit.Runner {
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
