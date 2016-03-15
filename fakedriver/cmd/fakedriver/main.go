package main

import (
	"flag"
	"os"

	cf_debug_server "github.com/cloudfoundry-incubator/cf-debug-server"
	cf_lager "github.com/cloudfoundry-incubator/cf-lager"

	"github.com/cloudfoundry-incubator/volman/fakedriver"
	"github.com/cloudfoundry-incubator/volman/voldriver/driverhttp"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
)

var atAddress = flag.String(
	"listenAddr",
	"0.0.0.0:9750",
	"host:port to serve volume management functions",
)

var driversPath = flag.String(
	"driversPath",
	"",
	"Path to directory where drivers are installed",
)

func init() {
	parseCommandLine()
}

func main() {
	withLogger, logTap := logger()
	withLogger.Info("started")
	defer withLogger.Info("ends")

	fakeDriverServer := createFakeDriverServer(withLogger, *atAddress, *driversPath)
	servers := grouper.Members{
		{"fakedriver-server", fakeDriverServer},
	}
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

func createFakeDriverServer(logger lager.Logger, atAddress string, driversPath string) ifrit.Runner {
	fileSystem := fakedriver.NewRealFileSystem()
	client := fakedriver.NewLocalDriver(&fileSystem)
	handler, err := driverhttp.NewHandler(logger, client)
	exitOnFailure(logger, err)
	return http_server.New(atAddress, handler)
}

func logger() (lager.Logger, *lager.ReconfigurableSink) {

	logger, reconfigurableSink := cf_lager.New("fakedriverServer")
	return logger, reconfigurableSink
}

func parseCommandLine() {
	cf_lager.AddFlags(flag.CommandLine)
	cf_debug_server.AddFlags(flag.CommandLine)
	flag.Parse()
}
