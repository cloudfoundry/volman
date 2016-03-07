package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cloudfoundry-incubator/volman/voldriver"
	flags "github.com/jessevdk/go-flags"
)

type InfoCommand struct {
	Info func() `short:"i" long:"info" description:"Print program information"`
}

func (x *InfoCommand) Execute(args []string) error {
	fakeDriver := NewLocalDriver()

	response, err := fakeDriver.Info(nil)
	if err != nil {
		return err
	}
	printJson(response)
	return nil
}

type MountCommand struct {
	Volume string `short:"v" long:"volume" description:"ID of the volume to mount"`
}

func (x *MountCommand) Execute(args []string) error {
	fakeDriver := NewLocalDriver()

	response, err := fakeDriver.Mount(nil, voldriver.MountRequest{x.Volume, ""})
	if err != nil {
		return err
	}
	printJson(response)
	return nil
}

type UnmountCommand struct {
	Volume string `short:"v" long:"volume" description:"ID of the volume Id to unmount"`
}

func (x *UnmountCommand) Execute(args []string) error {
	fakeDriver := NewLocalDriver()

	err := fakeDriver.Unmount(nil, voldriver.UnmountRequest{x.Volume})
	if err != nil {
		return err
	}
	printJson(struct{}{})
	return nil
}

type Options struct{}

func main() {
	var infoCmd InfoCommand
	var mountCmd MountCommand
	var unmountCmd UnmountCommand
	var options Options
	var parser = flags.NewParser(&options, flags.Default)

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

func printJson(response interface{}) {
	jsonBlob, err := json.Marshal(response)
	if err != nil {
		panic("Error Marshaling the driver")
	}
	fmt.Println(string(jsonBlob))
}
