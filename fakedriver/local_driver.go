package fakedriver

import (
	"errors"
	"fmt"
	"os"

	"strings"

	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/pivotal-golang/lager"
)

const RootDir = "_fakedriver/"

type LocalDriver struct { // see voldriver.resources.go
	volumes    map[string]*volume
	fileSystem FileSystem
}

type volume struct {
	volumeID   string
	mountpoint string
}

func NewLocalDriver(fileSystem FileSystem) *LocalDriver {
	return &LocalDriver{
		volumes:    map[string]*volume{},
		fileSystem: fileSystem,
	}
}

func (d *LocalDriver) Info(logger lager.Logger) (voldriver.InfoResponse, error) {
	return voldriver.InfoResponse{
		Name: "fakedriver",
		Path: "/fake/path",
	}, nil
}

func (d *LocalDriver) Mount(logger lager.Logger, mountRequest voldriver.MountRequest) voldriver.MountResponse {
	logger = logger.Session("mount", lager.Data{"volume": mountRequest.Name})

	var vol *volume
	var ok bool
	if vol, ok = d.volumes[mountRequest.Name]; !ok {
		return voldriver.MountResponse{Err: fmt.Sprintf("Volume '%s' must be created before being mounted", mountRequest.Name)}
	}

	mountPath := d.mountPath(vol.volumeID)

	logger.Info("mounting-volume", lager.Data{"id": vol.volumeID, "mountpoint": mountPath})
	err := d.fileSystem.MkdirAll(mountPath, 0777)
	if err != nil {
		logger.Error("failed-creating-mountpoint", err)
		return voldriver.MountResponse{Err: fmt.Sprintf("Error mounting volume: %s", err.Error())}
	}

	vol.mountpoint = mountPath

	mountResponse := voldriver.MountResponse{Mountpoint: mountPath}
	return mountResponse
}

func (d *LocalDriver) Unmount(logger lager.Logger, unmountRequest voldriver.UnmountRequest) voldriver.ErrorResponse {
	logger = logger.Session("unmount", lager.Data{"volume": unmountRequest.Name})

	mountPath, err := d.get(unmountRequest.Name)
	if err != nil {
		logger.Error("failed-no-such-volume-found", err, lager.Data{"mountpoint": mountPath})

		return voldriver.ErrorResponse{Err: fmt.Sprintf("Volume '%s' not found", unmountRequest.Name)}
	}

	if mountPath == "" {
		errText := "Volume not previously mounted"
		logger.Error("failed-mountpoint-not-assigned", errors.New(errText))
		return voldriver.ErrorResponse{Err: errText}
	}

	exists, err := d.exists(mountPath)
	if err != nil {
		logger.Error("failed-retrieving-mount-info", err, lager.Data{"mountpoint": mountPath})
		return voldriver.ErrorResponse{Err: "Error establishing whether volume exists"}
	}

	if !exists {
		errText := fmt.Sprintf("Volume %s does not exist, nothing to do!", unmountRequest.Name)
		logger.Error("failed-mountpoint-not-found", errors.New(errText))
		return voldriver.ErrorResponse{Err: errText}
	}

	logger.Info("removing-volume-path", lager.Data{"mountpoint": mountPath})
	err = d.fileSystem.RemoveAll(mountPath)
	if err != nil {
		logger.Error("failed-removing-mount-path", err)
		return voldriver.ErrorResponse{Err: fmt.Sprintf("Failed removing mount path: %s", err)}
	}
	logger.Info("unmounted-volume")

	volume := d.volumes[unmountRequest.Name]
	volume.mountpoint = ""
	return voldriver.ErrorResponse{}
}

func (d *LocalDriver) Create(logger lager.Logger, createRequest voldriver.CreateRequest) voldriver.ErrorResponse {
	logger = logger.Session("create")
	if id, ok := createRequest.Opts["volume_id"]; ok {
		logger.Info("creating-volume", lager.Data{"volume_name": createRequest.Name, "volume_id": id})

		if v, ok := d.volumes[createRequest.Name]; ok {
			// If a volume with the given name already exists, no-op unless the opts are different
			if v.volumeID != id {
				logger.Info("duplicate-volume", lager.Data{"volume_name": createRequest.Name})
				return voldriver.ErrorResponse{Err: fmt.Sprintf("Volume '%s' already exists with a different volume ID", createRequest.Name)}
			}

			return voldriver.ErrorResponse{}
		}

		d.volumes[createRequest.Name] = &volume{volumeID: id.(string)}
		return voldriver.ErrorResponse{}
	}

	logger.Info("missing-volume-id", lager.Data{"volume_name": createRequest.Name})
	return voldriver.ErrorResponse{Err: "Missing mandatory 'volume_id' field in 'Opts'"}
}

func (d *LocalDriver) Get(logger lager.Logger, getRequest voldriver.GetRequest) voldriver.GetResponse {
	mountpoint, err := d.get(getRequest.Name)
	if err != nil {
		return voldriver.GetResponse{Err: err.Error()}
	}

	return voldriver.GetResponse{Volume: voldriver.VolumeInfo{Name: getRequest.Name, Mountpoint: mountpoint}}
}

func (d *LocalDriver) get(volumeName string) (string, error) {
	if vol, ok := d.volumes[volumeName]; ok {
		return vol.mountpoint, nil
	}

	return "", errors.New("Volume not found")
}

func (d *LocalDriver) exists(path string) (bool, error) {
	_, err := d.fileSystem.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func (d *LocalDriver) mountPath(volumeId string) string {
	tmpDir := d.fileSystem.TempDir()
	if !strings.HasSuffix(tmpDir, "/") {
		tmpDir = fmt.Sprintf("%s/", tmpDir)
	}

	return tmpDir + RootDir + volumeId
}
