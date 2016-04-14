package certification

import (
	"encoding/json"
	"fmt"
	"github.com/tedsuo/ifrit/ginkgomon"
	"io/ioutil"
	"os/exec"
)

type VolmanConfig struct {
	ServerPort         int    `json:"ServerPort"`
	DebugServerAddress string `json:"DebugServerAddress"`
	DriversPath        string `json:"DriversPath"`
	ListenAddress      string `json:"ListenAddress"`
}

type VolmanFixture struct {
	Runner *ginkgomon.Runner
	Config VolmanConfig `json:"VolmanConfig"`
}

type DriverConfig struct {
	Name          string `json:"Name"`
	Path          string `json:"Path"`
	ServerPort    int    `json:"ServerPort"`
	ListenAddress string `json:"ListenAddress"`
}

type VolumeData struct {
	Name   string                 `json:"Name"`
	Config map[string]interface{} `json:"Config"`
}

type DriverFixture struct {
	Runner *ginkgomon.Runner

	Config DriverConfig `json:"DriverConfig"`

	VolumeData VolumeData `json:"VolumeData"`
}

type CertificationFixture struct {
	PathToVolman  string        `json:"PathToVolman"`
	VolmanFixture VolmanFixture `json:"VolmanFixture"`

	PathToDriver  string        `json:"PathToDriver"`
	MountDir      string        `json:"MountDir"`
	Transport     string        `json:"Transport"`
	DriverFixture DriverFixture `json:"DriverFixture"`
}

func LoadCertificationFixture(fileName string) (CertificationFixture, error) {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return CertificationFixture{}, err
	}

	certificationFixture := CertificationFixture{}
	err = json.Unmarshal(bytes, &certificationFixture)
	if err != nil {
		return CertificationFixture{}, err
	}

	return certificationFixture, nil
}

func SaveCertificationFixture(fixture CertificationFixture, fileName string) error {
	bytes, err := json.MarshalIndent(fixture, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fileName, bytes, 0666)
}

func CreateRunners(certifcationFixture *CertificationFixture) {
	CreateVolmanFixtureRunner(certifcationFixture)
	CreateDriverFixtureRunner(certifcationFixture)
}

func CreateVolmanFixtureRunner(certifcationFixture *CertificationFixture) {
	//certifcationFixturre.VolmanFixture.Config.ServerPort = 8750 + GinkgoParallelNode()
	//certifcationFixtue.VolmanFixture.Config.DebugServerAddress = fmt.Sprintf("0.0.0.0:%d", 8850+GinkgoParallelNode())

	certifcationFixture.VolmanFixture.Runner = ginkgomon.New(ginkgomon.Config{
		Name: "volman",
		Command: exec.Command(
			certifcationFixture.PathToVolman,
			"-listenAddr", certifcationFixture.VolmanFixture.Config.ListenAddress,
			"-debugAddr", certifcationFixture.VolmanFixture.Config.DebugServerAddress,
			"-driversPath", certifcationFixture.VolmanFixture.Config.DriversPath,
		),
		StartCheck: "volman.started",
	})
}

func CreateDriverFixtureRunner(certifcationFixture *CertificationFixture) {
	//switch certifcationFixture.Transport  {
	//case "tcp":
	//	certifcationFixture.DriverFixture.Config.ServerPort = 9650 + 100 + GinkgoParallelNode()
	//case "tcp-json":
	//	certifcationFixture.DriverFixture.Config.ServerPort = 9650 + GinkgoParallelNode()
	//case "unix":
	//	certifcationFixture.ListenAddress = path.Join(certifcationFixture.DriverFixture.Config.Path, "fakedriver.sock")
	//}

	certifcationFixture.DriverFixture.Runner = ginkgomon.New(ginkgomon.Config{
		Name: fmt.Sprintf("fakedriver%sServer", certifcationFixture.Transport),
		Command: exec.Command(
			certifcationFixture.PathToDriver,
			"-listenAddr", certifcationFixture.DriverFixture.Config.ListenAddress,
			"-transport", certifcationFixture.Transport,
			"-mountDir", certifcationFixture.MountDir,
			"-driversPath", certifcationFixture.DriverFixture.Config.Path,
		),
		StartCheck: "fakedriverServer.started",
	})
}
