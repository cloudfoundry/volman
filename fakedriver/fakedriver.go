package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry-incubator/volman"
	flags "github.com/jessevdk/go-flags"
)

type InfoCommand struct {
	Info func() `short:"i" long:"info" description:"Print program information"`
}

func (x *InfoCommand) Execute(args []string) error {
	driverInfo := volman.DriverInfo{
		Name: "fakedriver",
		Path: "/fake/path",
	}

	jsonBlob, err := json.Marshal(driverInfo)
	if err != nil {
		panic("Error Marshaling the driver")
	}
	fmt.Println(string(jsonBlob))

	return nil
}

type MountCommand struct {
	Mount func() `short:"m" long:"mount" description:"Mount a volume Id to a path"`
}

func (x *MountCommand) Execute(args []string) error {

	tmpDriversPath, err := ioutil.TempDir("", "fakeDriverMountPoint")
	mountPoint := volman.MountPointResponse{tmpDriversPath}

	jsonBlob, err := json.Marshal(mountPoint)
	if err != nil {
		panic("Error Marshaling the mount point")
	}
	fmt.Println(string(jsonBlob))

	return nil
}

type Options struct{}

func main() {
	var infoCmd InfoCommand
	var mountCmd MountCommand
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
	_, err := parser.Parse()

	if err != nil {
		panic(err)
		os.Exit(1)
	}
}
