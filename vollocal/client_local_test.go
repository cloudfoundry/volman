package vollocal_test

import (
	"bytes"
	"fmt"
	"io"

	"github.com/cloudfoundry-incubator/volman"
	_ "github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/cloudfoundry-incubator/volman/vollocal"
	"github.com/cloudfoundry-incubator/volman/volmanfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("Volman", func() {

	var fakeCmd *volmanfakes.FakeCmd
	var fakeExec *volmanfakes.FakeExec
	var validDriverInfoResponse io.ReadCloser
	var testLogger = lagertest.NewTestLogger("ClientTest")

	BeforeEach(func() {
		validDriverInfoResponse = stringCloser{bytes.NewBufferString("{\"Name\":\"fakedriver\",\"Path\":\"somePath\"}")}
	})

	Context("has no drivers in location", func() {
		BeforeEach(func() {
			client = vollocal.NewLocalClient(fmt.Sprintf("%s/tmp", defaultPluginsDirectory))
		})

		It("should report empty list of drivers", func() {
			drivers, err := client.ListDrivers(testLogger)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(drivers.Drivers)).To(Equal(0))
		})

	})

	Context("has driver in location", func() {
		BeforeEach(func() {
			client = vollocal.NewLocalClient(defaultPluginsDirectory)
		})

		It("should report list of drivers", func() {
			drivers, err := client.ListDrivers(testLogger)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(drivers.Drivers)).ToNot(Equal(0))
		})
		It("should report only one fakedriver", func() {
			drivers, err := client.ListDrivers(testLogger)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(drivers.Drivers)).To(Equal(1))
			Expect(drivers.Drivers[0].Name).To(Equal("fakedriver"))
		})

	})

	Context("when given valid driver path", func() {
		var driverName string
		BeforeEach(func() {
			fakeExec = new(volmanfakes.FakeExec)
			fakeCmd = new(volmanfakes.FakeCmd)
			fakeExec.CommandReturns(fakeCmd)
			client = vollocal.NewLocalClientWithExec(defaultPluginsDirectory, fakeExec)
			driverName = "fakedriver"
		})

		It("should be able to mount", func() {
			var validDriverMountResponse = stringCloser{bytes.NewBufferString("{\"Path\":\"/MountPoint\"}")}
			var stdOutResponses = [...]io.ReadCloser{validDriverInfoResponse, validDriverMountResponse}
			makeCommandReturnStdOutResponsesInOrder(fakeCmd, stdOutResponses)

			volumeId := "fake-volume"
			config := "Here is some config!"

			mountPath, err := client.Mount(testLogger, driverName, volumeId, config)
			Expect(err).NotTo(HaveOccurred())
			Expect(mountPath).NotTo(Equal(""))
		})

		It("should not be able to mount if mount fails", func() {
			var invalidDriverMountResponse = errCloser{bytes.NewBufferString("")}
			var stdOutResponses = [...]io.ReadCloser{validDriverInfoResponse, invalidDriverMountResponse}
			makeCommandReturnStdOutResponsesInOrder(fakeCmd, stdOutResponses)

			volumeId := "fake-volume"
			config := "Here is some config!"

			_, err := client.Mount(testLogger, driverName, volumeId, config)
			Expect(err).To(HaveOccurred())
		})

		It("should be able to unmount", func() {
			var validDriverUnmountResponse = stringCloser{bytes.NewBufferString("{}")}
			var stdOutResponses = [...]io.ReadCloser{validDriverInfoResponse, validDriverUnmountResponse}
			makeCommandReturnStdOutResponsesInOrder(fakeCmd, stdOutResponses)

			volumeId := "fake-volume"

			err := client.Unmount(testLogger, driverName, volumeId)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when unable to list drivers", func() {
			BeforeEach(func() {
				fakeCmd.StdoutPipeReturns(nil, fmt.Errorf("failure"))
			})

			It("should not be able to list drivers", func() {
				_, err := client.ListDrivers(testLogger)
				Expect(err).To(HaveOccurred())
			})

			It("should not be able to mount", func() {

				volumeId := "fake-volume"
				config := "Here is some config!"
				_, err := client.Mount(testLogger, driverName, volumeId, config)
				Expect(err).To(HaveOccurred())
			})

			It("should not be able to unmount", func() {
				volumeId := "fake-volume"

				err := client.Unmount(testLogger, driverName, volumeId)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when given invalid driver", func() {
			BeforeEach(func() {
				driverName = "does-not-exist"
			})

			It("should not be able to mount", func() {
				var validDriverMountResponse = stringCloser{bytes.NewBufferString("{\"Path\":\"/MountPoint\"}")}
				var stdOutResponses = [...]io.ReadCloser{validDriverInfoResponse, validDriverMountResponse}
				makeCommandReturnStdOutResponsesInOrder(fakeCmd, stdOutResponses)

				volumeId := "fake-volume"
				config := "Here is some config!"
				_, err := client.Mount(testLogger, driverName, volumeId, config)
				Expect(err).To(HaveOccurred())
			})

			It("should not be able to unmount", func() {
				var validDriverMountResponse = stringCloser{bytes.NewBufferString("{}")}
				var stdOutResponses = [...]io.ReadCloser{validDriverInfoResponse, validDriverMountResponse}
				makeCommandReturnStdOutResponsesInOrder(fakeCmd, stdOutResponses)

				volumeId := "fake-volume"

				err := client.Unmount(testLogger, driverName, volumeId)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})

func makeCommandReturnStdOutResponsesInOrder(fakeCmd *volmanfakes.FakeCmd, stdOutResponses []io.ReadCloser) {
	calls := 0
	fakeCmd.StdoutPipeStub = func() (io.ReadCloser, error) {
		defer func() { calls++ }()
		return stdOutResponses[calls], nil
	}
}

func whenListDriversIsRan() (volman.ListDriversResponse, error) {
	testLogger := lagertest.NewTestLogger("ClientTest")
	return client.ListDrivers(testLogger)
}
