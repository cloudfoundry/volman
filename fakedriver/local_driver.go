package main

import (
	"fmt"
	"os"

	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/pivotal-golang/lager"
)

type localDriver struct { // see voldriver.resources.go
	rootDir string
}

func NewLocalDriver() *localDriver {
	return &localDriver{"_fakedriver/"}
}

func (d *localDriver) Info(logger lager.Logger) (voldriver.InfoResponse, error) {
	return voldriver.InfoResponse{
		Name: "fakedriver",
		Path: "/fake/path",
	}, nil
}

func (d *localDriver) Mount(logger lager.Logger, mountRequest voldriver.MountRequest) (voldriver.MountResponse, error) {
	mountPath := os.TempDir() + d.rootDir + mountRequest.VolumeId
	err := os.MkdirAll(mountPath, 0777)
	if err != nil {
		fmt.Printf("Error creating volume %s", err.Error())
		return voldriver.MountResponse{}, fmt.Errorf("Error creating volume %s", err.Error())
	}
	mountPoint := voldriver.MountResponse{mountPath}
	return mountPoint, nil
}

func (d *localDriver) Unmount(logger lager.Logger, unmountRequest voldriver.UnmountRequest) error {
	logger = logger.Session("FakeDriver")
	logger.Info("start")
	defer logger.Info("end")

	f, err := os.Create("/tmp/fakedriver-unmount")
	if err != nil {
		panic("cant create file")
	}
	defer f.Close()

	mountPath := os.TempDir() + d.rootDir + unmountRequest.VolumeId
	exists, err := exists(mountPath)
	if err != nil {
		f.WriteString("Error establishing if volume exists")
		logger.Error("Error establishing if volume exists", err)
		return fmt.Errorf("Error establishing if volume exists")
	}
	if !exists {
		f.WriteString(fmt.Sprintf("Volume %s does not exist, nothing to do!", unmountRequest.VolumeId))
		logger.Error(fmt.Sprintf("Volume %s does not exist, nothing to do!", unmountRequest.VolumeId), err)
		return fmt.Errorf("Volume %s does not exist, nothing to do!", unmountRequest.VolumeId)
	} else {
		err := os.RemoveAll(mountPath)
		if err != nil {
			f.WriteString(fmt.Sprintf("Unexpected error removing mount path %s", unmountRequest.VolumeId))
			logger.Error(fmt.Sprintf("Unexpected error removing mount path %s", unmountRequest.VolumeId), err)
			return fmt.Errorf("Unexpected error removing mount path %s", unmountRequest.VolumeId)
		}
	}
	return nil
}
