package main

import (
	"fmt"
	"os"
	"time"

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

	f, _ := d.openLog()
	defer f.Close()

	mountPath := d.mountPath(mountRequest.VolumeId)

	d.writeLog(f, "Mounting volume %s", mountRequest.VolumeId)
	d.writeLog(f, "Creating volume path %s", mountPath)
	err := os.MkdirAll(mountPath, 0777)
	if err != nil {
		fmt.Printf("Error creating volume %s", err.Error())
		d.writeLog(f, "Error creating volume %s", err.Error())
		return voldriver.MountResponse{}, fmt.Errorf("Error creating volume %s", err.Error())
	}
	mountPoint := voldriver.MountResponse{mountPath}
	return mountPoint, nil
}

func (d *localDriver) Unmount(logger lager.Logger, unmountRequest voldriver.UnmountRequest) error {

	f, _ := d.openLog()
	defer f.Close()

	mountPath := d.mountPath(unmountRequest.VolumeId)

	exists, err := exists(mountPath)
	if err != nil {
		d.writeLog(f, "Error establishing if volume exists")
		return fmt.Errorf("Error establishing if volume exists")
	}
	if !exists {
		d.writeLog(f, "Volume %s does not exist, nothing to do!", unmountRequest.VolumeId)
		return fmt.Errorf("Volume %s does not exist, nothing to do!", unmountRequest.VolumeId)
	} else {
		d.writeLog(f, "Removing volume path %s", mountPath)
		err := os.RemoveAll(mountPath)
		if err != nil {
			d.writeLog(f, "Unexpected error removing mount path %s", unmountRequest.VolumeId)
			return fmt.Errorf("Unexpected error removing mount path %s", unmountRequest.VolumeId)
		}
		d.writeLog(f, "Unmounted volume %s", unmountRequest.VolumeId)
	}
	return nil
}

func (d *localDriver) mountPath(volumeId string) string {

	tmpDir := os.TempDir()
	if !strings.HasSuffix(tmpDir, "/") {
		tmpDir = fmt.Sprintf("%s/", tmpDir)
	}

	return tmpDir + d.rootDir + volumeId
}

func (d *localDriver) openLog() (*os.File, error) {
	f, err := os.OpenFile(d.logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(fmt.Sprintf("Can't create fakedriver log file %s", d.logFile))
	}
	return f, nil
}

func (d *localDriver) writeLog(f *os.File, msg string, args ...string) error {
	t := time.Now()
	_, err := f.WriteString(fmt.Sprintf("[%s] "+msg+"\n", t.Format(time.RFC3339), args))
	return err
}
