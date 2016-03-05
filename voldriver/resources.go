package voldriver

import (
	"github.com/pivotal-golang/lager"
)

type InfoResponse struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

//go:generate counterfeiter -o ../volmanfakes/fake_driver_client.go . Driver

type Driver interface {
	Info(logger lager.Logger) (InfoResponse, error)

	/**
	 * Mounts a volume
	 *
	 * If successful your CLI implementation is expected to write the MountResponse to stdout
	 * return nil for error and exit with 0
	 * If unsuccessful your CLI implementation is expected to write nothing to stdout and
	 * exit with 1
	 */
	Mount(logger lager.Logger, mountRequest MountRequest) (MountResponse, error)

	/**
	 * Unmounts a volume
	 *
	 * If successful your CLI implementation is expected to return nil and exits with code 0
	 * If unsuccessful your CLI implementation is expected to write nothing to stdout and
	 * exit with code 1
	 */
	Unmount(logger lager.Logger, unmountRequest UnmountRequest) error
}

type MountRequest struct {
	VolumeId string `json:"volumeId"`
	Config   string `json:"config"`
}

type MountResponse struct {
	Path string `json:"path"`
}

type UnmountRequest struct {
	VolumeId string `json:"volumeId"`
}
