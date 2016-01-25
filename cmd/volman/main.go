package main

import (
	"flag"
	"fmt"
	"net/http"

	// cf_debug_server "github.com/cloudfoundry-incubator/cf-debug-server"
	// cf_lager "github.com/cloudfoundry-incubator/cf-lager"
)

// var communicationTimeout = flag.Duration(
// 	"communicationTimeout",
// 	10*time.Second,
// 	"Timeout applied to all HTTP requests.",
// )

func main() {
	// cf_debug_server.AddFlags(flag.CommandLine)
	// cf_lager.AddFlags(flag.CommandLine)
	flag.Parse()
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
	//cf_http.Initialize(*communicationTimeout)
}

// func initializeVolmanServer(runner volmantypes.Runner) ifrit.Runner {
// 	return http_server.New(*listenAddr, handlers.New(runner))
// }

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}
