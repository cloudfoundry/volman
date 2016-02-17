package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cloudfoundry-incubator/volman"
	flags "github.com/jessevdk/go-flags"
)

type AddCommand struct {
	Info func() `short:"i" long:"info" description:"Print program information"`
}

func (x *AddCommand) Execute(args []string) error {
	driver := volman.Driver{
		Name:    "FakeDriver",
		Version: "0.0.1",
	}

	jsonBlob, err := json.Marshal(driver)
	if err != nil {
		panic("Error Marshaling the driver")
	}
	fmt.Println(string(jsonBlob))

	return nil
}

type Options struct{}

func main() {
	var addCommand AddCommand
	var options Options
	var parser = flags.NewParser(&options, flags.Default)

	parser.AddCommand("info",
		"Print Info",
		"The info command print the driver name and version.",
		&addCommand)
	_, err := parser.Parse()

	if err != nil {
		panic(err)
		os.Exit(1)
	}
}
