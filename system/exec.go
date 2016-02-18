package system

import (
	"io"
	"os/exec"
)

//go:generate counterfeiter -o ../volmanfakes/fake_cmd.go . Cmd

type Cmd interface {
	Start() error
	StdoutPipe() (io.ReadCloser, error)
	Wait() error
}

//go:generate counterfeiter -o ../volmanfakes/fake_exec.go . Exec

type Exec interface {
	Command(name string, arg ...string) Cmd
}

type SystemExec struct{}

func (_ *SystemExec) Command(name string, arg ...string) Cmd {
	return exec.Command(name, arg...)
}
