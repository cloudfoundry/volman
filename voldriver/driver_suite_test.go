package voldriver_test

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry-incubator/volman"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var client volman.Manager
var fakeDriverPath string

func TestDriver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Driver Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error

	fakeDriverPath, err = gexec.Build("github.com/cloudfoundry-incubator/volman/fakedriver", "-race")
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

// testing support types:

type errCloser struct{ io.Reader }

func (errCloser) Close() error                     { return nil }
func (errCloser) Read(p []byte) (n int, err error) { return 0, fmt.Errorf("any") }

type stringCloser struct{ io.Reader }

func (stringCloser) Close() error { return nil }
