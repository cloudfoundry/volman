package vollocal_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/cloudfoundry-incubator/volman"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	"testing"
)

var client volman.Manager

var defaultPluginsDirectory string

var fakeDriverPath string
var localDriverServerPort int
var debugServerAddress string
var localDriverProcess ifrit.Process
var localDriverRunner *ginkgomon.Runner

var tmpDriversPath string

var fakedriverServerPath string

var unixRunner *ginkgomon.Runner
var fakedriverUnixServerProcess ifrit.Process
var socketPath string

func TestDriver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Volman Local Client Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error

	fakeDriverPath, err = gexec.Build("github.com/cloudfoundry-incubator/volman/fakedriver/cmd/fakedriver", "-race")
	Expect(err).NotTo(HaveOccurred())
	return []byte(fakeDriverPath)
}, func(pathsByte []byte) {
	path := string(pathsByte)
	fakeDriverPath = strings.Split(path, ",")[0]
})

var _ = BeforeEach(func() {
	var err error
	tmpDriversPath, err = ioutil.TempDir("", "driversPath")
	Expect(err).NotTo(HaveOccurred())

	defaultPluginsDirectory, err = ioutil.TempDir(os.TempDir(), "clienttest")
	Expect(err).ShouldNot(HaveOccurred())

	localDriverServerPort = 9750 + GinkgoParallelNode()

	debugServerAddress = fmt.Sprintf("0.0.0.0:%d", 9850+GinkgoParallelNode())
	localDriverRunner = ginkgomon.New(ginkgomon.Config{
		Name: "fakedriverServer",
		Command: exec.Command(
			fakeDriverPath,
			"-listenAddr", fmt.Sprintf("0.0.0.0:%d", localDriverServerPort),
			"-debugAddr", debugServerAddress,
			"-driversPath", defaultPluginsDirectory,
		),
		StartCheck: "fakedriverServer.started",
	})
	socketPath = path.Join(defaultPluginsDirectory, "fakedriver.sock")
	unixRunner = ginkgomon.New(ginkgomon.Config{
		Name: "fakedriverUnixServer",
		Command: exec.Command(
			fakeDriverPath,
			"-listenAddr", socketPath,
			"-transport", "unix",
		),
		StartCheck: "fakedriverUnixServer.started",
	})

})

var _ = AfterEach(func() {
	ginkgomon.Kill(localDriverProcess)
})

var _ = SynchronizedAfterSuite(func() {

}, func() {
	gexec.CleanupBuildArtifacts()
})

// testing support types:

type errCloser struct{ io.Reader }

func (errCloser) Close() error                     { return nil }
func (errCloser) Read(p []byte) (n int, err error) { return 0, fmt.Errorf("any") }

type stringCloser struct{ io.Reader }

func (stringCloser) Close() error { return nil }
