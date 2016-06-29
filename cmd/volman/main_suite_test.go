package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var (
	binaryPath string
)

func TestVolman(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Volman Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error
	binaryPath, err = gexec.Build("github.com/cloudfoundry-incubator/volman/cmd/volman", "-race")
	Expect(err).NotTo(HaveOccurred())

	return []byte(binaryPath)
}, func(bytes []byte) {
	binaryPath = string(bytes)
})
