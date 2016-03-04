package voldriver

import (
	"bufio"
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

func NewCliInvoker(useExec system.Exec, executable string, args ...string) CliInvoker {
	invoker := CliInvoker{useExec, executable, nil}
	invoker.UseCmd = invoker.UseExec.Command(executable, args...)
	return invoker
}

func (invoker *CliInvoker) InvokeDriver(logger lager.Logger, output interface{}) error {
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
	if output == nil {
		return invoker.DriverError(logger, err, "decoding JSON")
	}

	r := bufio.NewReader(stdout)
	line, _, _ := r.ReadLine()
	fmt.Printf("------.....>%s", line)

	decoder := json.NewDecoder(stdout)

	err = decoder.Decode(&output)
	if err != nil {
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
