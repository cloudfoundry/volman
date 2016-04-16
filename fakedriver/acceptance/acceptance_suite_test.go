package acceptance_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/volman/certification"
	"github.com/onsi/gomega/gexec"
)

var (
	volmanPath         string
	volmanServerPort   int
	debugServerAddress string

	driverPath           string
	driverServerPort     int
	driverServerPortJson int

	mountDir       string
	tmpDriversPath string

	socketPath          string
	fixturesFuncMapFunc func() map[string]func() (certification.VolmanFixture, certification.DriverFixture)
)

func TestVolman(t *testing.T) {
	// these integration tests can take a bit, especially under load;
	// 1 second is too harsh
	SetDefaultEventuallyTimeout(10 * time.Second)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Fakedriver Certification Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	vPath, err := gexec.Build("github.com/cloudfoundry-incubator/volman/cmd/volman", "-race")
	Expect(err).NotTo(HaveOccurred())

	dPath, err := gexec.Build("github.com/cloudfoundry-incubator/volman/fakedriver/cmd/fakedriver", "-race")
	Expect(err).NotTo(HaveOccurred())

	return []byte(strings.Join([]string{vPath, dPath}, ","))
}, func(pathsByte []byte) {
	path := string(pathsByte)
	volmanPath = strings.Split(path, ",")[0]
	driverPath = strings.Split(path, ",")[1]

	createDefaultCertificationFixtures([]string{"tcp", "tcp-json", "unix"})
})

var _ = BeforeEach(func() {
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	//gexec.CleanupBuildArtifacts()
})

func GetOrCreateFixturesPath() (string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	fixturesPath := filepath.Join(workingDir, "..", "..", "fixtures", "auto")

	err = os.MkdirAll(fixturesPath, 0700)
	if err != nil {
		return "", err
	}

	return fixturesPath, nil
}

func createDefaultCertificationFixtures(transports []string) {
	var err error

	tmpDriversPath, err = ioutil.TempDir(os.TempDir(), "volman-cert-test")
	Expect(err).ShouldNot(HaveOccurred())

	mountDir, err = ioutil.TempDir("", "mountDir")
	Expect(err).NotTo(HaveOccurred())

	driverServerPort = 9750 + GinkgoParallelNode()
	driverServerPortJson = driverServerPort + 100
	volmanServerPort = 8750 + GinkgoParallelNode()

	fixturesPath, err := GetOrCreateFixturesPath()
	Expect(err).NotTo(HaveOccurred())

	ports := []int{driverServerPort, driverServerPortJson, 0}
	for i, transport := range transports {
		certificationFixture := certification.CertificationFixture{
			VolmanFixture: certification.VolmanFixture{
				PathToVolman: volmanPath,
				Config: certification.VolmanConfig{
					ServerPort:         volmanServerPort,
					DebugServerAddress: debugServerAddress,
					DriversPath:        tmpDriversPath,
					ListenAddress:      fmt.Sprintf("0.0.0.0:%d", volmanServerPort),
				},
			},
			DriverFixture: certification.DriverFixture{
				PathToDriver: driverPath,
				MountDir:     mountDir,
				Transport:    transport,
				Config: certification.DriverConfig{
					Name:          "fakedriver",
					Path:          tmpDriversPath,
					ServerPort:    ports[i],
					ListenAddress: fmt.Sprintf("0.0.0.0:%d", driverServerPort),
				},
				VolumeData: certification.VolumeData{
					Name: "fake-volume-name",
					Config: map[string]interface{}{
						"volume_id": "fake-volume-name",
					},
				},
			},
		}

		fileName := filepath.Join(fixturesPath, fmt.Sprintf("certification-%s.json", transport))
		err := certification.SaveCertificationFixture(certificationFixture, fileName)
		Expect(err).NotTo(HaveOccurred())
	}
}
