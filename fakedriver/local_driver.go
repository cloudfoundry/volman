package fakedriver

import (
	"errors"
	"fmt"
	"os"

	"strings"

	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/pivotal-golang/lager"
)

const RootDir = "_volumes/"

type LocalDriver struct { // see voldriver.resources.go
	volumes    map[string]*volume
	fileSystem FileSystem
	mountDir   string
}

type volume struct {
	volumeID   string
	mountpoint string
}

func NewLocalDriver(fileSystem FileSystem, mountDir string) *LocalDriver {
	return &LocalDriver{
		volumes:    map[string]*volume{},
		fileSystem: fileSystem,
		mountDir:   mountDir,
	}
}

func (d *LocalDriver) Activate(logger lager.Logger) voldriver.ActivateResponse {
	return voldriver.ActivateResponse{
		Implements: []string{"VolumeDriver"},
	}
}

func (d *LocalDriver) Create(logger lager.Logger, createRequest voldriver.CreateRequest) voldriver.ErrorResponse {
	logger = logger.Session("create")
	var ok bool
	var id interface{}

	if createRequest.Name == "" {
		return voldriver.ErrorResponse{Err: "Missing mandatory 'volume_name'"}
	}

	if id, ok = createRequest.Opts["volume_id"]; !ok {
		logger.Info("missing-volume-id", lager.Data{"volume_name": createRequest.Name})
		return voldriver.ErrorResponse{Err: "Missing mandatory 'volume_id' field in 'Opts'"}
	}

	var existingVolume *volume
	if existingVolume, ok = d.volumes[createRequest.Name]; !ok {
		logger.Info("creating-volume", lager.Data{"volume_name": createRequest.Name, "volume_id": id.(string)})
		d.volumes[createRequest.Name] = &volume{volumeID: id.(string)}
		return voldriver.ErrorResponse{}
	}

	// If a volume with the given name already exists, no-op unless the opts are different
	if existingVolume.volumeID != id {
		logger.Info("duplicate-volume", lager.Data{"volume_name": createRequest.Name})
		return voldriver.ErrorResponse{Err: fmt.Sprintf("Volume '%s' already exists with a different volume ID", createRequest.Name)}
	}

	return voldriver.ErrorResponse{}
}

func (d *LocalDriver) Mount(logger lager.Logger, mountRequest voldriver.MountRequest) voldriver.MountResponse {
	logger = logger.Session("mount", lager.Data{"volume": mountRequest.Name})

	if mountRequest.Name == "" {
		return voldriver.MountResponse{Err: "Missing mandatory 'volume_name'"}
	}

	var vol *volume
	var ok bool
	if vol, ok = d.volumes[mountRequest.Name]; !ok {
		return voldriver.MountResponse{Err: fmt.Sprintf("Volume '%s' must be created before being mounted", mountRequest.Name)}
	}

	mountPath := d.mountPath(logger, vol.volumeID)

	logger.Info("mounting-volume", lager.Data{"id": vol.volumeID, "mountpoint": mountPath})
	err := d.fileSystem.MkdirAll(mountPath, os.ModePerm)
	if err != nil {
		logger.Error("failed-creating-mountpoint", err)
		return voldriver.MountResponse{Err: fmt.Sprintf("Error mounting volume: %s", err.Error())}
	}

	vol.mountpoint = mountPath

	mountResponse := voldriver.MountResponse{Mountpoint: mountPath}
	return mountResponse
}

func (d *LocalDriver) Path(logger lager.Logger, pathRequest voldriver.PathRequest) voldriver.PathResponse {
	logger = logger.Session("path", lager.Data{"volume": pathRequest.Name})

	if pathRequest.Name == "" {
		return voldriver.PathResponse{Err: "Missing mandatory 'volume_name'"}
	}

	mountPath, err := d.get(logger, pathRequest.Name)
	if err != nil {
		logger.Error("failed-no-such-volume-found", err, lager.Data{"mountpoint": mountPath})

		return voldriver.PathResponse{Err: fmt.Sprintf("Volume '%s' not found", pathRequest.Name)}
	}

	if mountPath == "" {
		errText := "Volume not previously mounted"
		logger.Error("failed-mountpoint-not-assigned", errors.New(errText))
		return voldriver.PathResponse{Err: errText}
	}

	return voldriver.PathResponse{Mountpoint: mountPath}
}

func (d *LocalDriver) Unmount(logger lager.Logger, unmountRequest voldriver.UnmountRequest) voldriver.ErrorResponse {
	logger = logger.Session("unmount", lager.Data{"volume": unmountRequest.Name})

	if unmountRequest.Name == "" {
		return voldriver.ErrorResponse{Err: "Missing mandatory 'volume_name'"}
	}

	mountPath, err := d.get(logger, unmountRequest.Name)
	if err != nil {
		logger.Error("failed-no-such-volume-found", err, lager.Data{"mountpoint": mountPath})

		return voldriver.ErrorResponse{Err: fmt.Sprintf("Volume '%s' not found", unmountRequest.Name)}
	}

	if mountPath == "" {
		errText := "Volume not previously mounted"
		logger.Error("failed-mountpoint-not-assigned", errors.New(errText))
		return voldriver.ErrorResponse{Err: errText}
	}

	return d.unmount(logger, unmountRequest.Name, mountPath)
}

func (d *LocalDriver) Remove(logger lager.Logger, removeRequest voldriver.RemoveRequest) voldriver.ErrorResponse {
	logger = logger.Session("remove", lager.Data{"volume": removeRequest})
	logger.Info("start")
	defer logger.Info("end")

	if removeRequest.Name == "" {
		return voldriver.ErrorResponse{Err: "Missing mandatory 'volume_name'"}
	}

	var response voldriver.ErrorResponse
	var vol *volume
	var exists bool
	if vol, exists = d.volumes[removeRequest.Name]; !exists {
		logger.Error("failed-volume-removal", fmt.Errorf(fmt.Sprintf("Volume %s not found", removeRequest.Name)))
		return voldriver.ErrorResponse{fmt.Sprintf("Volume '%s' not found", removeRequest.Name)}
	}

	if vol.mountpoint != "" {
		response = d.unmount(logger, removeRequest.Name, vol.mountpoint)
		if response.Err != "" {
			return response
		}
	}

	logger.Info("removing-volume", lager.Data{"name": removeRequest.Name})
	delete(d.volumes, removeRequest.Name)
	return voldriver.ErrorResponse{}
}

func (d *LocalDriver) Get(logger lager.Logger, getRequest voldriver.GetRequest) voldriver.GetResponse {
	mountpoint, err := d.get(logger, getRequest.Name)
	if err != nil {
		return voldriver.GetResponse{Err: err.Error()}
	}

	return voldriver.GetResponse{Volume: voldriver.VolumeInfo{Name: getRequest.Name, Mountpoint: mountpoint}}
}

func (d *LocalDriver) get(logger lager.Logger, volumeName string) (string, error) {
	if vol, ok := d.volumes[volumeName]; ok {
		logger.Info("getting-volume", lager.Data{"name": volumeName})
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

func (d *LocalDriver) mountPath(logger lager.Logger, volumeId string) string {
	dir, err := d.fileSystem.Abs(d.mountDir)
	if err != nil {
		logger.Fatal("error getting path to executable", err)
	}

	if !strings.HasSuffix(dir, "/") {
		dir = fmt.Sprintf("%s/", dir)
	}

	return dir + RootDir + volumeId
}

func (d *LocalDriver) unmount(logger lager.Logger, name string, mountPath string) voldriver.ErrorResponse {
	exists, err := d.exists(mountPath)
	if err != nil {
		logger.Error("failed-retrieving-mount-info", err, lager.Data{"mountpoint": mountPath})
		return voldriver.ErrorResponse{Err: "Error establishing whether volume exists"}
	}

	if !exists {
		errText := fmt.Sprintf("Volume %s does not exist (path: %s), nothing to do!", name, mountPath)
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

	volume := d.volumes[name]
	volume.mountpoint = ""

	return voldriver.ErrorResponse{}
}
