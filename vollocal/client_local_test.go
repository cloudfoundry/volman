package vollocal_test

import (
	"fmt"
	"bytes"
	"io"
	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/vollocal"
	_"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/cloudfoundry-incubator/volman/volmanfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"

)

var _ = Describe("Volman", func() {

	var fakeCmd *volmanfakes.FakeCmd
	var fakeExec *volmanfakes.FakeExec
	var validDriverInfoResponse io.ReadCloser

	Context("has no drivers in location", func() {
		BeforeEach(func() {
			client = &vollocal.LocalClient{DriversPath: fmt.Sprintf("%s/tmp",defaultPluginsDirectory)}
		})

		It("should report empty list of drivers", func() {
			testLogger := lagertest.NewTestLogger("ClientTest")
			drivers, err := client.ListDrivers(testLogger)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(drivers.Drivers)).To(Equal(0))
		})

		
	})
	Context("has driver in location", func() {
		BeforeEach(func() {
			client = &vollocal.LocalClient{DriversPath: defaultPluginsDirectory}
		})

		It("should report list of drivers", func() {
			testLogger := lagertest.NewTestLogger("ClientTest")
			drivers, err := client.ListDrivers(testLogger)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(drivers.Drivers)).ToNot(Equal(0))
		})
		It("should report only one fakedriver", func() {
			testLogger := lagertest.NewTestLogger("ClientTest")
			drivers, err := client.ListDrivers(testLogger)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(drivers.Drivers)).To(Equal(1))
			Expect(drivers.Drivers[0].Name).To(Equal("fakedriver"))
		})

	})




Context("when given valid driver path", func() {
		BeforeEach(func() {
			fakeExec = new(volmanfakes.FakeExec)
			fakeCmd = new(volmanfakes.FakeCmd)
			fakeExec.CommandReturns(fakeCmd)
			client = &vollocal.LocalClient{DriversPath: defaultPluginsDirectory}
		})

		It("should be able to mount", func() {
			var validDriverMountResponse = stringCloser{bytes.NewBufferString("{\"Path\":\"/MountPoint\"}")}
			var stdOutResponses = [...]io.ReadCloser{validDriverMountResponse, validDriverInfoResponse}

			calls := 0
			fakeCmd.StdoutPipeStub = func() (io.ReadCloser, error) {
				defer func() { calls++ }()
				return stdOutResponses[calls], nil
			}
			testLogger := lagertest.NewTestLogger("ClientTest")
			volumeId := "fake-volume"
			config := "Here is some config!"
			mountPath, err := client.Mount(testLogger, "fakedriver", volumeId, config)
			Expect(err).NotTo(HaveOccurred())
			Expect(mountPath).NotTo(Equal(""))
		})
	})
})
func whenListDriversIsRan() (volman.ListDriversResponse, error) {
	testLogger := lagertest.NewTestLogger("ClientTest")
	return client.ListDrivers(testLogger)
}
