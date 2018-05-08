package voldiscoverers_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var defaultPluginsDirectory string
var secondPluginsDirectory string

var _ = BeforeEach(func() {
	var err error

	defaultPluginsDirectory, err = ioutil.TempDir(os.TempDir(), "clienttest")
	Expect(err).ShouldNot(HaveOccurred())

	secondPluginsDirectory, err = ioutil.TempDir(os.TempDir(), "clienttest2")
	Expect(err).ShouldNot(HaveOccurred())
})

func TestVoldiscoverers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Voldiscoverers Suite")
}
