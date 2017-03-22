package main

import (
	"flag"
	"os"

	"path/filepath"

	cf_lager "code.cloudfoundry.org/cflager"
	cf_debug_server "code.cloudfoundry.org/debugserver"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/loggregator_v2"
	"code.cloudfoundry.org/volman/volhttp"
	"code.cloudfoundry.org/volman/vollocal"
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

var driverPaths = flag.String(
	"volmanDriverPaths",
	"",
	"Path to the directory where drivers can be discovered.  Multiple paths may be specified using the OS-specific path separator; e.g. /path/to/somewhere:/path/to/somewhere-else",
)

var loggregatorUseV2API = flag.Bool(
	"loggregatorUseV2Api",
	false,
	"boolean to use the v2 loggregator api or fall back to dropsonde",
)

var loggregatorApiPort = flag.Int(
	"loggregatorApiPort",
	0,
	"api port",
)

var loggregatorCaCertPath = flag.String(
	"loggregatorCACertPath",
	"",
	"ca certificate path",
)
var loggregatorCertPath = flag.String(
	"loggregatorCertPath",
	"",
	"Certificate path for loggregator",
)
var loggregatorKeyPath = flag.String(
	"loggregatorKeyPath",
	"",
	"Loggregator key path",
)
var loggregatorJobDeployment = flag.String(
	"loggregatorJobDeployment",
	"",
	"Loggreagaotr Job deployment",
)
var loggregatorJobName = flag.String(
	"loggregatorJobName",
	"",
	"Loggregator Job Name",
)
var loggregatorJobIP = flag.String(
	"loggregatorJobIP",
	"",
	"Loggregator Job IP",
)
var loggregatorJobIndex = flag.String(
	"loggregatorJobIndex",
	"",
	"Loggregator Job Index",
)
var loggregatorJobOrigin = flag.String(
	"loggregatorJobOrigin",
	"",
	"Loggregator Job Origin",
)
var loggregatorDropsondePort = flag.Int(
	"loggregatorDropsondePort",
	0,
	"Loggregator dropsonde port",
)

func init() {
	// no command line parsing can happen here in go 1.6
}

func main() {
	parseCommandLine()
	logger, logSink := cf_lager.New("volman")
	defer logger.Info("ends")

	metronConfig := &loggregator_v2.MetronConfig{
		UseV2API:      *loggregatorUseV2API,
		APIPort:       *loggregatorApiPort,
		CACertPath:    *loggregatorCaCertPath,
		CertPath:      *loggregatorCertPath,
		KeyPath:       *loggregatorKeyPath,
		JobDeployment: *loggregatorJobDeployment,
		JobName:       *loggregatorJobName,
		JobIndex:      *loggregatorJobIndex,
		JobIP:         *loggregatorJobIP,
		JobOrigin:     *loggregatorJobOrigin,
		DropsondePort: *loggregatorDropsondePort,
	}

	metronClient, err := loggregator_v2.NewClient(logger, *metronConfig)
	if err != nil {
		logger.Error("failed-to-initialize-metron-client", err)
		os.Exit(1)
	}

	servers := createVolmanServer(logger, *atAddress, *driverPaths, metronClient)

	if dbgAddr := cf_debug_server.DebugAddress(flag.CommandLine); dbgAddr != "" {
		servers = append(grouper.Members{
			{"debug-server", cf_debug_server.Runner(dbgAddr, logSink)},
		}, servers...)
	}
	process := ifrit.Invoke(processRunnerFor(servers))
	logger.Info("started")
	untilTerminated(logger, process)
}

func exitOnFailure(logger lager.Logger, err error) {
	if err != nil {
		logger.Error("fatal-err-aborting", err)
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

func createVolmanServer(logger lager.Logger, atAddress string, driverPaths string, metronClient loggregator_v2.Client) grouper.Members {
	if driverPaths == "" {
		panic("'-volmanDriverPaths' must be provided")
	}

	cfg := vollocal.NewDriverConfig()
	cfg.DriverPaths = filepath.SplitList(driverPaths)
	client, runner := vollocal.NewServer(logger, metronClient, cfg)

	handler, err := volhttp.NewHandler(logger, client)
	exitOnFailure(logger, err)

	return grouper.Members{
		{"driver-syncer", runner},
		{"http-server", http_server.New(atAddress, handler)},
	}
}

func parseCommandLine() {
	cf_lager.AddFlags(flag.CommandLine)
	cf_debug_server.AddFlags(flag.CommandLine)
	flag.Parse()
}
