package fakedriver

import (
	"fmt"
	"os"

	"strings"

	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/pivotal-golang/lager"
)

type localDriver struct { // see voldriver.resources.go
	rootDir string
	logFile string
}

func NewLocalDriver() *localDriver {
	return &localDriver{"_fakedriver/", "/tmp/fakedriver.log"}
}

func (d *localDriver) Info(logger lager.Logger) (voldriver.InfoResponse, error) {
	return voldriver.InfoResponse{
		Name: "fakedriver",
		Path: "/fake/path",
	}, nil
}

func (d *localDriver) Mount(logger lager.Logger, mountRequest voldriver.MountRequest) (voldriver.MountResponse, error) {

	mountPath := d.mountPath(mountRequest.VolumeId)

	logger.Info(fmt.Sprintf("Mounting volume %s", mountRequest.VolumeId))
	logger.Info(fmt.Sprintf("Creating volume path %s", mountPath))
	err := os.MkdirAll(mountPath, 0777)
	if err != nil {
		logger.Info(fmt.Sprintf("Error creating volume %s", err.Error()))
		return voldriver.MountResponse{}, fmt.Errorf("Error creating volume %s", err.Error())
	}
	mountPoint := voldriver.MountResponse{mountPath}
	return mountPoint, nil
}

func (d *localDriver) Unmount(logger lager.Logger, unmountRequest voldriver.UnmountRequest) error {

	mountPath := d.mountPath(unmountRequest.VolumeId)

	exists, err := exists(mountPath)
	if err != nil {
		logger.Info(fmt.Sprintf("Error establishing if volume exists"))
		return fmt.Errorf("Error establishing if volume exists")
	}
	if !exists {
		logger.Info(fmt.Sprintf("Volume %s does not exist, nothing to do!", unmountRequest.VolumeId))
		return fmt.Errorf("Volume %s does not exist, nothing to do!", unmountRequest.VolumeId)
	} else {
		logger.Info(fmt.Sprintf("Removing volume path %s", mountPath))
		err := os.RemoveAll(mountPath)
		if err != nil {
			logger.Info(fmt.Sprintf("Unexpected error removing mount path %s", unmountRequest.VolumeId))
			return fmt.Errorf("Unexpected error removing mount path %s", unmountRequest.VolumeId)
		}
		logger.Info(fmt.Sprintf("Unmounted volume %s", unmountRequest.VolumeId))
	}
	return nil
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func (d *localDriver) mountPath(volumeId string) string {

	tmpDir := os.TempDir()
	if !strings.HasSuffix(tmpDir, "/") {
		tmpDir = fmt.Sprintf("%s/", tmpDir)
	}

	return tmpDir + d.rootDir + volumeId
}
