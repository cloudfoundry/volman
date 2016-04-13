package certification

import (
	"fmt"
	"github.com/tedsuo/ifrit/ginkgomon"
	"os/exec"
)

type VolmanFixtureCreator interface {
	Create(pathToVolman string) VolmanFixture
}

func NewVolmanFixtureCreator(serverPort int, debugServerAddress string, driversPath string) *VolmanFixture {
	return &VolmanFixture{
		Runner: nil,
		Config: VolmanConfig{
			ServerPort:         serverPort,
			DebugServerAddress: debugServerAddress,
			DriversPath:        driversPath,
		},
	}
}

func (vf *VolmanFixture) Create(pathToVolman string) VolmanFixture {
	runner := ginkgomon.New(ginkgomon.Config{
		Name: "volman",
		Command: exec.Command(
			pathToVolman,
			"-listenAddr", fmt.Sprintf("0.0.0.0:%d", vf.Config.ServerPort),
			"-debugAddr", vf.Config.DebugServerAddress,
			"-driversPath", vf.Config.DriversPath,
		),
		StartCheck: "volman.started",
	})

	return VolmanFixture{
		Runner: runner,
		Config: vf.Config,
	}
}
