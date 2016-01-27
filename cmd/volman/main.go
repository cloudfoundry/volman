package main

import (
	"flag"
	"os"

	cf_debug_server "github.com/cloudfoundry-incubator/cf-debug-server"
	cf_lager "github.com/cloudfoundry-incubator/cf-lager"
	. "github.com/cloudfoundry-incubator/volman/delegate"
	"github.com/cloudfoundry-incubator/volman/handlers"
	"github.com/pivotal-golang/lager"
	. "github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
)

var atAddress = flag.String(
	"listenAddr",
	"0.0.0.0:8750",
	"host:port to serve volume management functions",
)

func main() {
	parseCommandLine()
	withLogger, logTap := logger()
	servers := grouper.Members{}
	servers = createAndAppendServer(*atAddress, servers)
	servers = createAndAppendDebugServer(cf_debug_server.DebugAddress(flag.CommandLine), logTap, servers)
	untilTerminated(Invoke(processRunnerFor(servers)), withLogger)
}

func exitOnFailure(err error) {
	if err != nil {
		os.Exit(1)
	}
}

func untilTerminated(process Process, logger lager.Logger) {
	logger.Info("started")
	err := <-process.Wait()
	exitOnFailure(err)
	logger.Info("exited")
}

func processRunnerFor(servers grouper.Members) Runner {
	return sigmon.New(grouper.NewOrdered(os.Interrupt, servers))
}

func createAndAppendDebugServer(atDebugAddress string, logTap *lager.ReconfigurableSink, servers grouper.Members) grouper.Members {
	if atDebugAddress == "" {
		return servers
	}

	return append(grouper.Members{
		{"debug-server", cf_debug_server.Runner(atDebugAddress, logTap)},
	}, servers...)
}

func createAndAppendServer(atAddress string, servers grouper.Members) grouper.Members {
	handler, err := handlers.Generate()
	exitOnFailure(err)
	server := http_server.New(atAddress, handler)
	return append(grouper.Members{
		{"volman-server", server},
	}, servers...)
}

func logger() (lager.Logger, *lager.ReconfigurableSink) {

	logger, reconfigurableSink := cf_lager.New("volman")
	return logger, reconfigurableSink
}

func parseCommandLine() {
	cf_lager.AddFlags(flag.CommandLine)
	cf_debug_server.AddFlags(flag.CommandLine)
	flag.Parse()
}
