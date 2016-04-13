package certification

import (
	"github.com/tedsuo/ifrit/ginkgomon"
	"os/exec"
)

type DriverFixtureCreator interface {
	Create(pathToDriver string, mountDir string, transport string, listenAddress string) DriverFixture

	CreateRequestExample(methodName string, parameters map[string]string) error
	Reset() error
}

func NewDriverFixtureCreator(driverName string, driverPath string, driverServerPort int, volumeName string, volumeConfig map[string]interface{}) *DriverFixture {
	return &DriverFixture{
		Runner: nil,

		Config: DriverConfig{
			Name:       driverName,
			Path:       driverPath,
			ServerPort: driverServerPort,
		},
		VolumeData: VolumeData{
			Name:   volumeName,
			Config: volumeConfig,
		},
	}
}

func (df *DriverFixture) Create(pathToDriver string, mountDir string, transport string, listenAddress string) DriverFixture {
	runner := ginkgomon.New(ginkgomon.Config{
		Name: "fakedriverServer",
		Command: exec.Command(
			pathToDriver,
			"-listenAddr", listenAddress,
			"-mountDir", mountDir,
			"-transport", transport,
			"-driversPath", df.Config.Path,
		),
		StartCheck: "fakedriverServer.started",
	})

	return DriverFixture{
		Runner:     runner,
		Config:     df.Config,
		VolumeData: df.VolumeData,
	}
}

func (df *DriverFixture) CreateRequestExample(methodName string, parameters map[string]string) error {
	return nil
}

func (df *DriverFixture) Reset() error {
	return nil
}
