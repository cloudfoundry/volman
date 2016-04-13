package certification

import (
	"encoding/json"
	"github.com/tedsuo/ifrit"
	"io/ioutil"
)

type VolmanConfig struct {
	ServerPort         int    `json:"ServerPort"`
	DebugServerAddress string `json:"DebugServerAddress"`
	DriversPath        string `json:"DriversPath"`
}

type VolmanFixture struct {
	Runner ifrit.Runner
	Config VolmanConfig `json:"VolmanConfig"`
}

type DriverConfig struct {
	Name       string `json:"Name"`
	Path       string `json:"Path"`
	ServerPort int    `json:"ServerPort"`
}

type VolumeData struct {
	Name   string                 `json:"Name"`
	Config map[string]interface{} `json:"Config"`
}

type DriverFixture struct {
	Runner ifrit.Runner

	Config DriverConfig `json:"DriverConfig"`

	VolumeData VolumeData `json:"VolumeData"`
}

type CertificationFixture struct {
	PathToVolman  string        `json:"PathToVolman"`
	VolmanFixture VolmanFixture `json:"VolmanFixture"`

	PathToDriver  string        `json:"PathToDriver"`
	MountDir      string        `json:"MountDir"`
	Transport     string        `json:"Transport"`
	ListenAddress string        `json:"ListenAddress"`
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
