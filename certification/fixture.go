package certification

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

type CertificationFixture struct {
	VolmanDriverPath  string                  `json:"volman_driver_path"`
	DriverName        string                  `json:"driver_name"`
	ResetDriverScript string                  `json:"reset_driver_script"`
	CreateConfig      voldriver.CreateRequest `json:"create_config"`
}

func NewCertificationFixture(volmanDriverPath string, driverName string, resetDriverScript string, createConfig voldriver.CreateRequest) *CertificationFixture {
	return &CertificationFixture{
		VolmanDriverPath:  volmanDriverPath,
		DriverName:        driverName,
		ResetDriverScript: resetDriverScript,
		CreateConfig:      createConfig,
	}
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

func (cf *CertificationFixture) CreateVolmanRunner() ifrit.Runner {
	volmanBinPath := os.Getenv("VOLMAN_PATH")
	return ginkgomon.New(ginkgomon.Config{
		Name: "volman",
		Command: exec.Command(
			volmanBinPath,
			"-listenAddr", fmt.Sprintf("0.0.0.0:%d", 8750),
			"-driversPath", cf.VolmanDriverPath,
		),
		StartCheck: "volman.started",
	})

}
