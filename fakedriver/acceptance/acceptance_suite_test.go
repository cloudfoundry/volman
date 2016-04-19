package acceptance_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var (
	volmanPath         string
	volmanServerPort   int
	debugServerAddress string

	driverPath           string
	driverServerPort     int
	driverServerPortJson int
	driverProcess        ifrit.Process

	mountDir       string
	tmpDriversPath string

	socketPath string
)

func TestVolman(t *testing.T) {
	// these integration tests can take a bit, especially under load;
	// 1 second is too harsh
	SetDefaultEventuallyTimeout(10 * time.Second)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Fakedriver Certification Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {

	return []byte{}
}, func(pathsByte []byte) {
	createDefaultCertificationFixtures([]string{"tcp"})

})

var _ = SynchronizedAfterSuite(func() {}, func() {
	ginkgomon.Kill(driverProcess)
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

	mountDir, err = ioutil.TempDir("", "mountDir")
	Expect(err).NotTo(HaveOccurred())

	// TODO get this from env variable os.Getenv("FIXTURES_PATH")

	// fixtures := certification.NewCertificationFixture("/tmp/builds/volman", "/tmp/plugins", "fakedriver", "/tmp/scripts/resetdriver.sh", voldriver.CreateRequest{})

	// fixturesPath := "/tmp/fixtures"
	// fileName := filepath.Join(fixturesPath, "certification.json")
	// err = certification.SaveCertificationFixture(*fixtures, fileName)
	// Expect(err).NotTo(HaveOccurred())

}
