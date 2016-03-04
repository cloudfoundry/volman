package voldriver_test

import (
	"bytes"
	"io"

	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/cloudfoundry-incubator/volman/volmanfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("DriverClientCli", func() {
	var client voldriver.Driver
	var fakeCmd *volmanfakes.FakeCmd
	var fakeExec *volmanfakes.FakeExec
	var validInfoResponseResponse = stringCloser{bytes.NewBufferString("")}
	var invalidInfoResponseResponse io.ReadCloser

	Context("when given invalid driver path", func() {
		BeforeEach(func() {
			fakeExec = new(volmanfakes.FakeExec)
			fakeCmd = new(volmanfakes.FakeCmd)
			fakeExec.CommandReturns(fakeCmd)
			client = &voldriver.DriverClientCli{fakeExec, fakeDriverPath, "SomeDriver"}
		})
		It("should error on get driver info", func() {
			var stdOutResponses = [...]io.ReadCloser{invalidInfoResponseResponse}

			fakeCmd.StdoutPipeStub = func() (io.ReadCloser, error) {
				return stdOutResponses[0], nil
			}

			testLogger := lagertest.NewTestLogger("ClientTest")

			_, err := client.Info(testLogger)
			Expect(err).To(HaveOccurred())

		})

		It("should not be able to mount", func() {
			fakeCmd.StdoutPipeReturns(errCloser{bytes.NewBufferString("")}, nil)
			testLogger := lagertest.NewTestLogger("ClientTest")

			volumeId := "fake-volume"
			config := "Here is some config!"

			mountResponse, err := client.Mount(testLogger, voldriver.MountRequest{VolumeId: volumeId, Config: config})
			Expect(err).To(HaveOccurred())
			Expect(mountResponse.Path).To(Equal(""))
		})

		It("should not be able to unmount", func() {
			// redundant test
		})

	})

	Context("when given valid driver path", func() {
		BeforeEach(func() {
			fakeExec = new(volmanfakes.FakeExec)
			fakeCmd = new(volmanfakes.FakeCmd)
			fakeExec.CommandReturns(fakeCmd)

			validInfoResponseResponse = stringCloser{bytes.NewBufferString("{\"Name\":\"fakedriver\",\"Path\":\"somePath\"}")}

			client = &voldriver.DriverClientCli{fakeExec, fakeDriverPath, "fakedriver"}
		})

		It("should not error on get driver info", func() {

			var stdOutResponses = [...]io.ReadCloser{validInfoResponseResponse}

			fakeCmd.StdoutPipeStub = func() (io.ReadCloser, error) {
				return stdOutResponses[0], nil
			}

			testLogger := lagertest.NewTestLogger("ClientTest")

			InfoResponse, err := client.Info(testLogger)
			Expect(err).NotTo(HaveOccurred())
			Expect(InfoResponse.Name).To(Equal("fakedriver"))
			Expect(InfoResponse.Path).To(Equal("somePath"))

		})

		It("should be able to mount", func() {
			var validDriverMountResponse = stringCloser{bytes.NewBufferString("{\"Path\":\"/MountPoint\"}")}
			var stdOutResponses = [...]io.ReadCloser{validDriverMountResponse}

			fakeCmd.StdoutPipeStub = func() (io.ReadCloser, error) {
				return stdOutResponses[0], nil
			}

			testLogger := lagertest.NewTestLogger("ClientTest")
			volumeId := "fake-volume"
			config := "Here is some config!"

			mountResponse, err := client.Mount(testLogger, voldriver.MountRequest{VolumeId: volumeId, Config: config})
			Expect(err).NotTo(HaveOccurred())
			Expect(mountResponse.Path).To(Equal("/MountPoint"))
		})

		It("should be able to unmount", func() {
			var validDriverUnmountResponse = stringCloser{bytes.NewBufferString("{}")}
			var stdOutResponses = [...]io.ReadCloser{validDriverUnmountResponse}

			fakeCmd.StdoutPipeStub = func() (io.ReadCloser, error) {
				return stdOutResponses[0], nil
			}

			testLogger := lagertest.NewTestLogger("ClientTest") // todo: move this to top level before
			volumeId := "fake-volume"

			err := client.Unmount(testLogger, voldriver.UnmountRequest{VolumeId: volumeId})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should not unmount a volume that doesnt exist", func() {
			var invalidDriverUnmountResponse = errCloser{bytes.NewBufferString("")}
			var stdOutResponses = [...]io.ReadCloser{invalidDriverUnmountResponse}

			fakeCmd.StdoutPipeStub = func() (io.ReadCloser, error) {
				return stdOutResponses[0], nil
			}

			testLogger := lagertest.NewTestLogger("ClientTest") // todo: move this to top level before
			volumeId := "fake-volume"

			err := client.Unmount(testLogger, voldriver.UnmountRequest{VolumeId: volumeId})
			Expect(err).To(HaveOccurred())
		})
	})
})
