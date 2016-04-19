package acceptance_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestVolman(t *testing.T) {
	SetDefaultEventuallyTimeout(10 * time.Second)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Fakedriver Certification Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	volmanPath, err := gexec.Build("github.com/cloudfoundry-incubator/volman/cmd/volman", "-race")
	Expect(err).NotTo(HaveOccurred())
	os.Setenv("VOLMAN_PATH", volmanPath)
	driverPath, err := gexec.Build("github.com/cloudfoundry-incubator/volman/fakedriver/cmd/fakedriver", "-race")
	Expect(err).NotTo(HaveOccurred())
	os.Setenv("DRIVER_PATH", driverPath)
	return []byte(strings.Join([]string{volmanPath, driverPath}, ","))
}, func(pathsByte []byte) {
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
	cmd := exec.Command("/bin/bash", "../scripts/stopdriver.sh")
	err := cmd.Run()
	Expect(err).NotTo(HaveOccurred())
})
