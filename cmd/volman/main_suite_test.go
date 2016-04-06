package main_test

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"testing"
	"time"

	//"os"
	"path"

	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var volmanPath string
var volmanServerPort int
var debugServerAddress string
var volmanRunner *ginkgomon.Runner

var driverPath string
var driverServerPort int
var debugServerAddress2 string
var driverRunner *ginkgomon.Runner
var unixDriverRunner *ginkgomon.Runner

var mountDir string
var tmpDriversPath string
var volumeName string
var socketPath string
var opts map[string]interface{}

func TestVolman(t *testing.T) {
	// these integration tests can take a bit, especially under load;
	// 1 second is too harsh
	SetDefaultEventuallyTimeout(10 * time.Second)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Volman Cmd Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error
	volmanPath, err = gexec.Build("github.com/cloudfoundry-incubator/volman/cmd/volman", "-race")
	Expect(err).NotTo(HaveOccurred())

	driverPath, err = gexec.Build("github.com/cloudfoundry-incubator/volman/fakedriver/cmd/fakedriver", "-race")
	Expect(err).NotTo(HaveOccurred())
	return []byte(strings.Join([]string{volmanPath, driverPath}, ","))
}, func(pathsByte []byte) {
	path := string(pathsByte)
	volmanPath = strings.Split(path, ",")[0]
	driverPath = strings.Split(path, ",")[1]
})

var _ = BeforeEach(func() {
	var err error

	tmpDriversPath, err = ioutil.TempDir(os.TempDir(), "volman-cert-test")
	Expect(err).ShouldNot(HaveOccurred())

	mountDir, err = ioutil.TempDir("", "mountDir")
	Expect(err).NotTo(HaveOccurred())

	driverServerPort = 9750 + GinkgoParallelNode()
	driverRunner = ginkgomon.New(ginkgomon.Config{
		Name: "fakedriverServer",
		Command: exec.Command(
			driverPath,
			"-listenAddr", fmt.Sprintf("0.0.0.0:%d", driverServerPort),
			"-mountDir", mountDir,
		),
		StartCheck: "fakedriverServer.started",
	})

	socketPath = path.Join(tmpDriversPath, "fakedriver.sock")
	unixDriverRunner = ginkgomon.New(ginkgomon.Config{
		Name: "fakedriverUnixServer",
		Command: exec.Command(
			driverPath,
			"-listenAddr", socketPath,
			"-transport", "unix",
			"-mountDir", mountDir,
		),
		StartCheck: "fakedriverUnixServer.started",
	})

	volmanServerPort = 8750 + GinkgoParallelNode()
	debugServerAddress = fmt.Sprintf("0.0.0.0:%d", 8850+GinkgoParallelNode())
	volmanRunner = ginkgomon.New(ginkgomon.Config{
		Name: "volman",
		Command: exec.Command(
			volmanPath,
			"-listenAddr", fmt.Sprintf("0.0.0.0:%d", volmanServerPort),
			"-debugAddr", debugServerAddress,
			"-driversPath", tmpDriversPath,
		),
		StartCheck: "volman.started",
	})
	volumeName = "fake-volume-name"
	opts = map[string]interface{}{"volume_id": volumeName}
})

var _ = SynchronizedAfterSuite(func() {

}, func() {
	gexec.CleanupBuildArtifacts()
})
