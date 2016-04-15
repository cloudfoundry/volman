package main

import (
	"flag"
	"os"

	cf_debug_server "github.com/cloudfoundry-incubator/cf-debug-server"
	cf_lager "github.com/cloudfoundry-incubator/cf-lager"
	"github.com/cloudfoundry-incubator/volman/volhttp"
	"github.com/cloudfoundry-incubator/volman/vollocal"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
)

var atAddress = flag.String(
	"listenAddr",
	"0.0.0.0:8750",
	"host:port to serve volume management functions",
)

var driversPath = flag.String(
	"driversPath",
	"",
	"Path to directory where drivers are installed",
)

func init() {
	// no command line parsing can happen here in go 1.6
}

func main() {
	parseCommandLine()
	withLogger, logTap := logger()
	withLogger.Info("started")
	defer withLogger.Info("ends")

	servers := createVolmanServer(withLogger, *atAddress, *driversPath)

	if dbgAddr := cf_debug_server.DebugAddress(flag.CommandLine); dbgAddr != "" {
		servers = append(grouper.Members{
			{"debug-server", cf_debug_server.Runner(dbgAddr, logTap)},
		}, servers...)
	}
	process := ifrit.Invoke(processRunnerFor(servers))
	untilTerminated(withLogger, process)
}

func exitOnFailure(logger lager.Logger, err error) {
	if err != nil {
		logger.Error("Fatal err.. aborting", err)
		panic(err.Error())
	}
}

func untilTerminated(logger lager.Logger, process ifrit.Process) {
	err := <-process.Wait()
	exitOnFailure(logger, err)
}

func processRunnerFor(servers grouper.Members) ifrit.Runner {
	return sigmon.New(grouper.NewOrdered(os.Interrupt, servers))
}

func createVolmanServer(logger lager.Logger, atAddress string, driversPath string) grouper.Members {
	if driversPath == "" {
		panic("'-driversPath' must be provided")
	}

	client, runner := vollocal.NewLocalClient(logger, driversPath)

	handler, err := volhttp.NewHandler(logger, client)
	exitOnFailure(logger, err)

	return grouper.Members{
		{"driver-syncer", runner},
		{"http-server", http_server.New(atAddress, handler)},
	}
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
