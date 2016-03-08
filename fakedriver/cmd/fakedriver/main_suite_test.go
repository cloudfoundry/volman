package main_test

import (
	"fmt"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var fakedriverServerPath string

var runner *ginkgomon.Runner
var fakedriverServerPort int
var fakedriverServerProcess ifrit.Process

var debugServerAddress string

func TestFakedriverServer(t *testing.T) {
	// these integration tests can take a bit, especially under load;
	// 1 second is too harsh
	SetDefaultEventuallyTimeout(10 * time.Second)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Fakedriver Server Cmd Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error
	fakedriverServerPath, err = gexec.Build("github.com/cloudfoundry-incubator/volman/fakedriver/cmd/fakedriver", "-race")
	Expect(err).NotTo(HaveOccurred())

	return []byte(fakedriverServerPath)
}, func(pathsByte []byte) {
	fakedriverServerPath = string(pathsByte)
})

var _ = BeforeEach(func() {

	fakedriverServerPort = 9750 + GinkgoParallelNode()
	debugServerAddress = fmt.Sprintf("0.0.0.0:%d", 9850+GinkgoParallelNode())
	runner = ginkgomon.New(ginkgomon.Config{
		Name: "fakedriverServer",
		Command: exec.Command(
			fakedriverServerPath,
			"-listenAddr", fmt.Sprintf("0.0.0.0:%d", fakedriverServerPort),
			"-debugAddr", debugServerAddress,
		),
		StartCheck: "fakedriverServer.started",
	})
})

var _ = AfterEach(func() {
	ginkgomon.Kill(fakedriverServerProcess)
})

var _ = SynchronizedAfterSuite(func() {

}, func() {
	gexec.CleanupBuildArtifacts()
})
