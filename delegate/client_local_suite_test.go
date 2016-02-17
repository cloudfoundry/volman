package delegate_test

import (
	"path/filepath"
	"strings"

	"github.com/cloudfoundry-incubator/volman"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var client volman.Client
var fakeDriverPath string

func TestLocalClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Local Client Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error

	fakeDriverPath, err = gexec.Build("github.com/cloudfoundry-incubator/volman/volmanfakes", "-race")
	Expect(err).NotTo(HaveOccurred())
	return []byte(strings.Join([]string{fakeDriverPath}, ","))
}, func(pathsByte []byte) {
	path := string(pathsByte)
	fakeDriverPath = filepath.Dir(strings.Split(path, ",")[0])
})

var _ = SynchronizedAfterSuite(func() {

}, func() {
	gexec.CleanupBuildArtifacts()
})
