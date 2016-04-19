package driverhttp_test

import (
	"fmt"
	"io"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	"os"
	"os/exec"
	"path"
	"testing"
)

var debugServerAddress string
var fakeDriverPath string

var fakedriverServerPort int
var fakedriverProcess ifrit.Process
var tcpRunner *ginkgomon.Runner

var unixRunner *ginkgomon.Runner
var fakedriverUnixServerProcess ifrit.Process
var socketPath string

func TestDriver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Driver Remote Client and Handlers Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error

	fakeDriverPath, err = gexec.Build("github.com/cloudfoundry-incubator/volman/fakedriver/cmd/fakedriver", "-race")
	Expect(err).NotTo(HaveOccurred())
	return []byte(fakeDriverPath)
}, func(pathsByte []byte) {
	fakeDriverPath = string(pathsByte)
})

var _ = BeforeEach(func() {

	tmpdir, err := ioutil.TempDir(os.TempDir(), "fake-driver-test")
	Expect(err).ShouldNot(HaveOccurred())

	socketPath = path.Join(tmpdir, "fakedriver.sock")

	unixRunner = ginkgomon.New(ginkgomon.Config{
		Name: "fakedriverUnixServer",
		Command: exec.Command(
			fakeDriverPath,
			"-listenAddr", socketPath,
			"-transport", "unix",
		),
		StartCheck: "fakedriverServer.started",
	})
})

var _ = AfterEach(func() {
	ginkgomon.Kill(fakedriverUnixServerProcess)
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
