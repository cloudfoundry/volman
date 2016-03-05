package voldriver

import (
	"encoding/json"
	"fmt"

	"github.com/cloudfoundry-incubator/volman/system"
	"github.com/pivotal-golang/lager"
)

type CliInvoker struct {
	UseExec    system.Exec
	executable string
	UseCmd     system.Cmd
}

func NewCliInvoker(useExec system.Exec, executable string) *CliInvoker {
	return &CliInvoker{useExec, executable, nil}
}

func (invoker *CliInvoker) Command(args ...string) {
	invoker.UseCmd = invoker.UseExec.Command(invoker.executable, args...)
}

func (invoker *CliInvoker) Execute(logger lager.Logger, output interface{}) error {
	if output == nil {
		return invoker.DriverError(logger, nil, "invoke-driver illegal argument: <output> must not be nil")
	}

	logger = logger.Session("invoke-driver")
	logger.Info("start")
	defer logger.Info("end")

	cmd := invoker.UseCmd
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return invoker.DriverError(logger, err, "fetching stdout")
	}

	if err := cmd.Start(); err != nil {
		return invoker.DriverError(logger, err, "starting")
	}

	if err := json.NewDecoder(stdout).Decode(&output); err != nil {
		return invoker.DriverError(logger, err, "decoding JSON")
	}

	if err := cmd.Wait(); err != nil {
		return invoker.DriverError(logger, err, "waiting")
	}
	return nil
}

func (invoker *CliInvoker) DriverError(logger lager.Logger, err error, specifics string) error {
	logger.Error(fmt.Sprintf("Error (%s) %s from driver at %s", err.Error(), specifics, invoker.executable), err)
	return err
}
