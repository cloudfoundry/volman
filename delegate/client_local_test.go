package delegate_test

import (
	"bytes"
	"fmt"
	"io"

	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/delegate"
	"github.com/cloudfoundry-incubator/volman/volmanfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("ListDrivers", func() {

	Context("has no drivers", func() {
		BeforeEach(func() {
			client = &delegate.LocalClient{}
		})

		It("should report empty list of drivers", func() {
			testLogger := lagertest.NewTestLogger("ClientTest")
			drivers, err := client.ListDrivers(testLogger)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(drivers.Drivers)).To(Equal(0))
		})
	})

	Context("has drivers", func() {
		var fakeCmd *volmanfakes.FakeCmd

		BeforeEach(func() {
			var fakeExec = new(volmanfakes.FakeExec)
			fakeCmd = new(volmanfakes.FakeCmd)
			fakeExec.CommandReturns(fakeCmd)

			client = &delegate.LocalClient{fakeDriverPath, fakeExec}
		})

		It("should error if the driver's stdout cannot be fetched", func() {
			fakeCmd.StdoutPipeReturns(nil, fmt.Errorf("StdoutPipe Error"))

			_, err := whenListDriversIsRan()

			Expect(err).Should(HaveOccurred())
		})

		It("should error if driver execution cannot be started", func() {
			fakeCmd.StartReturns(fmt.Errorf("Permission denied"))

			_, err := whenListDriversIsRan()

			Expect(err).Should(HaveOccurred())
		})

		It("should error if a driver's JSON is malformed", func() {
			fakeCmd.StdoutPipeReturns(errCloser{bytes.NewBufferString("")}, nil)

			_, err := whenListDriversIsRan()

			Expect(err).Should(HaveOccurred())

		})

		Context("that report valid JSON info", func() {
			BeforeEach(func() {
				fakeCmd.StdoutPipeReturns(stringCloser{bytes.NewBufferString("{\"Name\":\"SomeDriver\"}")}, nil)
			})
			It("should report list of drivers", func() {
				drivers, err := whenListDriversIsRan()
				Expect(err).ShouldNot(HaveOccurred())
				Expect(drivers.Drivers[0].Name).To(Equal("SomeDriver"))
			})

			It("should error if driver execution completes with failure", func() {
				fakeCmd.WaitReturns(fmt.Errorf("any"))

				_, err := whenListDriversIsRan()

				Expect(err).Should(HaveOccurred())
			})
		})
	})

})

func whenListDriversIsRan() (volman.ListDriversResponse, error) {
	testLogger := lagertest.NewTestLogger("ClientTest")
	return client.ListDrivers(testLogger)
}

type errCloser struct{ io.Reader }

func (errCloser) Close() error                     { return nil }
func (errCloser) Read(p []byte) (n int, err error) { return 0, fmt.Errorf("any") }

type stringCloser struct{ io.Reader }

func (stringCloser) Close() error { return nil }
