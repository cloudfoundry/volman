package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/voldriver"
	flags "github.com/jessevdk/go-flags"
)

var rootDir string

type InfoCommand struct {
	Info func() `short:"i" long:"info" description:"Print program information"`
}

func (x *InfoCommand) Execute(args []string) error {
	InfoResponse := voldriver.InfoResponse{
		Name: "fakedriver",
		Path: "/fake/path",
	}

	jsonBlob, err := json.Marshal(InfoResponse)
	if err != nil {
		panic("Error Marshaling the driver")
	}
	fmt.Println(string(jsonBlob))

	return nil
}

type MountCommand struct {
	Mount string `short:"v" long:"volume" description:"ID of the volume to mount"`
}

func (x *MountCommand) Execute(args []string) error {

	mountPath := os.TempDir() + rootDir + x.Mount
	err := os.MkdirAll(mountPath, 0777)
	if err != nil {
		fmt.Printf("Error creating volume %s", err.Error())
		return nil
	}
	mountPoint := volman.MountResponse{mountPath}

	jsonBlob, err := json.Marshal(mountPoint)
	if err != nil {
		panic("Error marshaling mount response")
	}
	fmt.Println(string(jsonBlob))

	return nil
}

type UnmountCommand struct {
	Unmount string `short:"v" long:"volume" description:"ID of the volume Id to unmount"`
}

func (x *UnmountCommand) Execute(args []string) error {

	mountPath := os.TempDir() + rootDir + x.Unmount
	exists, err := exists(mountPath)
	if err != nil {
		return fmt.Errorf("Error establishing if volume exists")
	}
	if !exists {
		return fmt.Errorf("Volume %s does not exist, nothing to do!", x.Unmount)
	}

	fmt.Println("{}")
	return nil
}

type Options struct{}

func main() {
	var infoCmd InfoCommand
	var mountCmd MountCommand
	var unmountCmd UnmountCommand
	var options Options
	var parser = flags.NewParser(&options, flags.Default)
	rootDir = "_fakedriver/"

	parser.AddCommand("info",
		"Print Info",
		"The info command print the driver name and version.",
		&infoCmd)
	parser.AddCommand("mount",
		"Mount Volume",
		"Mount a volume Id to a path - returning the path.",
		&mountCmd)
	parser.AddCommand("unmount",
		"Unnount Volume",
		"Unmount a volume Id",
		&unmountCmd)
	_, err := parser.Parse()

	if err != nil {
		//		panic(err)
		os.Exit(1)
	}
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
