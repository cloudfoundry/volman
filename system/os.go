package system

import "os"

//go:generate counterfeiter -o ../volmanfakes/fake_os.go . Os

type Os interface {
	Open(name string) (*os.File, error)
}

type SystemOs struct{}

func (_ *SystemOs) Open(name string) (*os.File, error) {
	return os.Open(name)
}
