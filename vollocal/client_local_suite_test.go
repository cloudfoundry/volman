package vollocal_test

import (
	"fmt"
	"io"

	"github.com/cloudfoundry-incubator/volman"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var client volman.Manager

var defaultPluginsDirectory string

func TestLocalClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Local Client Suite")
}

var _ = SynchronizedBeforeSuite(
	func() []byte {
		return []byte("")
	},
	func(pathsByte []byte) {
		defaultPluginsDirectory = "/tmp"
	},
)

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
