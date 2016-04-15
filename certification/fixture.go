package certification

import (
	"encoding/json"
	"io/ioutil"

	"github.com/tedsuo/ifrit/ginkgomon"
)

type VolmanConfig struct {
	ServerPort         int    `json:"ServerPort"`
	DebugServerAddress string `json:"DebugServerAddress"`
	DriversPath        string `json:"DriversPath"`
	ListenAddress      string `json:"ListenAddress"`
}

type VolmanFixture struct {
	Runner       *ginkgomon.Runner
	PathToVolman string       `json:"PathToVolman"`
	Config       VolmanConfig `json:"VolmanConfig"`
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

	MountDir     string       `json:"MountDir"`
	Transport    string       `json:"Transport"`
	PathToDriver string       `json:"PathToDriver"`
	Config       DriverConfig `json:"DriverConfig"`

	VolumeData VolumeData `json:"VolumeData"`
}

type CertificationFixture struct {
	VolmanFixture VolmanFixture `json:"VolmanFixture"`
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
	certifcationFixture.VolmanFixture.CreateRunner()
	certifcationFixture.DriverFixture.CreateRunner()
}
