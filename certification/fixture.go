package certification

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"

	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

type CertificationFixture struct {
	VolmanBinPath      string                  `json:"volman_bin_path"`
	VolmanDriverPath   string                  `json:"volman_driver_path"`
	DriverName         string                  `json:"driver_name"`
	ResetDriverScript  string                  `json:"reset_driver_script"`
	ValidCreateRequest voldriver.CreateRequest `json:"valid_create_request"`
}

func NewCertificationFixture(volmanBinPath string, volmanDriverPath string, driverName string, resetDriverScript string, validCreateRequest voldriver.CreateRequest) *CertificationFixture {
	return &CertificationFixture{
		VolmanBinPath:      volmanBinPath,
		VolmanDriverPath:   volmanDriverPath,
		DriverName:         driverName,
		ResetDriverScript:  resetDriverScript,
		ValidCreateRequest: validCreateRequest,
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
	return ginkgomon.New(ginkgomon.Config{
		Name: "volman",
		Command: exec.Command(
			cf.VolmanBinPath,
			"-listenAddr", fmt.Sprintf("0.0.0.0:%d", 8750),
			"-driversPath", cf.VolmanDriverPath,
		),
		StartCheck: "volman.started",
	})

}
