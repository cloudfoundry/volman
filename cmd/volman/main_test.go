package main_test

import (
	"github.com/cloudfoundry-incubator/volman/certification"
	. "github.com/onsi/ginkgo"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Fake Driver Certification", func() {
	certification.CertifiyWith("Fakedriver", func() (*ginkgomon.Runner, *ginkgomon.Runner, int, string, string, int) {
		return driverRunner, volmanRunner, volmanServerPort, debugServerAddress, tmpDriversPath, driverServerPort
	})
})
