package voldriver_test

import (
	"bytes"
	"io"

	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/cloudfoundry-incubator/volman/volmanfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"

)

var _ = Describe("DriverClientCli", func() {
	var client volman.DriverPlugin
	var fakeCmd *volmanfakes.FakeCmd
	var fakeExec *volmanfakes.FakeExec
	var validDriverInfoResponse io.ReadCloser

	// Context("when given invalid driver path", func() {
	// 	BeforeEach(func() {
	// 		fakeExec = new(volmanfakes.FakeExec)
	// 		fakeCmd = new(volmanfakes.FakeCmd)
	// 		fakeExec.CommandReturns(fakeCmd)

	// 		validDriverInfoResponse = stringCloser{bytes.NewBufferString("{\"Name\":\"SomeDriver\"}")}

	// 		client = &voldriver.DriverClientCli{fakeExec, fakeDriverPath}
	// 	})

		
	// 	It("should not be able to mount", func() {
	// 		fakeCmd.StdoutPipeReturns(errCloser{bytes.NewBufferString("")}, nil)
	// 		testLogger := lagertest.NewTestLogger("ClientTest")
	// 		driver := volman.Driver{
	// 			Name: "SomeDriver",
	// 		}
	// 		volumeId := "fake-volume"
	// 		config := "Here is some config!"

	// 		_, err := client.Mount(testLogger, driver, volumeId, config)
	// 		Expect(err).To(HaveOccurred())
	// 	})
	// })

	Context("when given valid driver path", func() {
		BeforeEach(func() {
			fakeExec = new(volmanfakes.FakeExec)
			fakeCmd = new(volmanfakes.FakeCmd)
			fakeExec.CommandReturns(fakeCmd)

			validDriverInfoResponse = stringCloser{bytes.NewBufferString("{\"Name\":\"SomeDriver\"}")}

			client = &voldriver.DriverClientCli{fakeExec,fakeDriverPath,"fakedriver"}
		})
		
		It("should not error on get driver info", func() {
			var validDriverMountResponse = stringCloser{bytes.NewBufferString("{\"Path\":\"/MountPoint\"}")}
			var stdOutResponses = [...]io.ReadCloser{validDriverMountResponse, validDriverInfoResponse}

			calls := 0
			fakeCmd.StdoutPipeStub = func() (io.ReadCloser, error) {
				defer func() { calls++ }()
				return stdOutResponses[calls], nil
			}

			testLogger := lagertest.NewTestLogger("ClientTest")
			
			_,err := client.Info(testLogger)
			Expect(err).NotTo(HaveOccurred())

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

			mountPath, err := client.Mount(testLogger, volumeId, config)
			Expect(err).NotTo(HaveOccurred())
			Expect(mountPath).To(Equal("/MountPoint"))
		})
	})
})
