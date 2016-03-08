package main_test

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var volmanPath string
var volmanServerPort int
var debugServerAddress string
var volmanProcess ifrit.Process
var runner *ginkgomon.Runner

var fakeDriverPath string
var fakedriverServerPort int
var debugServerAddress2 string
var fakedriverProcess ifrit.Process
var fakedriverRunner *ginkgomon.Runner

var tmpDriversPath string

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

	fakeDriverPath, err = gexec.Build("github.com/cloudfoundry-incubator/volman/fakedriver/cmd/fakedriver", "-race")
	Expect(err).NotTo(HaveOccurred())
	return []byte(strings.Join([]string{volmanPath, fakeDriverPath}, ","))
}, func(pathsByte []byte) {
	path := string(pathsByte)
	volmanPath = strings.Split(path, ",")[0]
	fakeDriverPath = strings.Split(path, ",")[1]
})

var _ = BeforeEach(func() {
	var err error
	tmpDriversPath, err = ioutil.TempDir("", "driversPath")
	Expect(err).NotTo(HaveOccurred())

	fakedriverServerPort = 9750 + GinkgoParallelNode()
	debugServerAddress2 = fmt.Sprintf("0.0.0.0:%d", 9850+GinkgoParallelNode())
	fakedriverRunner = ginkgomon.New(ginkgomon.Config{
		Name: "fakedriverServer",
		Command: exec.Command(
			fakeDriverPath,
			"-listenAddr", fmt.Sprintf("0.0.0.0:%d", fakedriverServerPort),
			"-debugAddr", debugServerAddress2,
		),
		StartCheck: "fakedriverServer.started",
	})

	volmanServerPort = 8750 + GinkgoParallelNode()
	debugServerAddress = fmt.Sprintf("0.0.0.0:%d", 8850+GinkgoParallelNode())
	runner = ginkgomon.New(ginkgomon.Config{
		Name: "volman",
		Command: exec.Command(
			volmanPath,
			"-listenAddr", fmt.Sprintf("0.0.0.0:%d", volmanServerPort),
			"-debugAddr", debugServerAddress,
			"-driversPath", tmpDriversPath,
		),
		StartCheck: "volman.started",
	})
})

var _ = AfterEach(func() {
	ginkgomon.Kill(volmanProcess)
	ginkgomon.Kill(fakedriverProcess)
})

var _ = SynchronizedAfterSuite(func() {

}, func() {
	gexec.CleanupBuildArtifacts()
})
