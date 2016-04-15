package certification

import (
	"fmt"
	"math/rand"
	"os/exec"
	"path/filepath"

	"github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"

	"github.com/tedsuo/ifrit/ginkgomon"
)

type DriverFixtureCreator interface {
	Create(pathToDriver string, mountDir string, transport string, listenAddress string) DriverFixture

	CreateRequestExample(methodName string, parameters map[string]string) error
	Reset() error

	CreateRunner()
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

func (df *DriverFixture) CreateRunner() {
	portOffset := rand.Intn(100)
	df.Config.ServerPort = df.Config.ServerPort + portOffset + GinkgoParallelNode()

	switch df.Transport {
	case "tcp":
		df.Config.ListenAddress = fmt.Sprintf("0.0.0.0:%d", df.Config.ServerPort)
	case "tcp-json":
		df.Config.ListenAddress = fmt.Sprintf("0.0.0.0:%d", df.Config.ServerPort+100)
	case "unix":
		df.Config.Name = fmt.Sprintf("fakedriver-%d", portOffset+GinkgoParallelNode())
		df.Config.ListenAddress = filepath.Join(df.Config.Path, fmt.Sprintf("%s.sock", df.Config.Name))
	default:
		panic(fmt.Sprintf("Invalid transport protocol: %s", df.Transport))
	}

	df.MountDir = fmt.Sprintf("%s-%d", df.MountDir, GinkgoParallelNode())

	df.Runner = ginkgomon.New(ginkgomon.Config{
		Name: fmt.Sprintf("fakedriver%sServer", df.Transport),
		Command: exec.Command(
			df.PathToDriver,
			"-listenAddr", df.Config.ListenAddress,
			"-transport", df.Transport,
			"-mountDir", df.MountDir,
			"-driversPath", df.Config.Path,
		),
		StartCheck: "fakedriverServer.started",
	})
}

func (df *DriverFixture) UpdateVolumeData() error {
	uuid, err := uuid.NewV4()
	if err != nil {
		return err
	}

	df.VolumeData.Name = fmt.Sprintf("fake-volume-name-%s", uuid.String())
	df.VolumeData.Config = map[string]interface{}{

		"volume_id": fmt.Sprintf("fake-volume-name-%s", uuid.String()),
	}

	return nil
}
