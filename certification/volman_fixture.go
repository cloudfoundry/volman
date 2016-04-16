package certification

import (
	"fmt"
	"math/rand"
	"os/exec"

	. "github.com/onsi/ginkgo"

	"github.com/tedsuo/ifrit/ginkgomon"
)

type VolmanFixtureCreator interface {
	Create(pathToVolman string) VolmanFixture
	CreateRunner()
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

func (vf *VolmanFixture) CreateRunner() {
	portOffset := rand.Intn(30)
	vf.Config.ServerPort = vf.Config.ServerPort + portOffset + GinkgoParallelNode()

	vf.Config.ListenAddress = fmt.Sprintf("0.0.0.0:%d", vf.Config.ServerPort)
	vf.Config.DebugServerAddress = fmt.Sprintf("0.0.0.0:%d", vf.Config.ServerPort+100)

	vf.Runner = ginkgomon.New(ginkgomon.Config{
		Name: "volman",
		Command: exec.Command(
			vf.PathToVolman,
			"-listenAddr", vf.Config.ListenAddress,
			"-debugAddr", vf.Config.DebugServerAddress,
			"-driversPath", vf.Config.DriversPath,
		),
		StartCheck: "volman.started",
	})
}
