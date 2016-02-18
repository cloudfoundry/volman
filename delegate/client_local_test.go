package delegate_test

import (
	"bytes"
	"fmt"
	"io"

	"github.com/cloudfoundry-incubator/volman/delegate"
	"github.com/cloudfoundry-incubator/volman/volmanfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("ListDrivers", func() {

	Context("volman has no drivers", func() {
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

	Context("volman has drivers", func() {
		var fakeCmd volmanfakes.FakeCmd

		BeforeEach(func() {
			var fakeExec = new(volmanfakes.FakeExec)
			fakeExec.CommandReturns(&fakeCmd)

			client = &delegate.LocalClient{fakeDriverPath, fakeExec}
		})

		It("should report list of drivers", func() {
			fakeCmd.StdoutPipeReturns(stringCloser{bytes.NewBufferString("{\"Name\":\"SomeDriver\"}")}, nil)
			testLogger := lagertest.NewTestLogger("ClientTest")
			drivers, err := client.ListDrivers(testLogger)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(drivers.Drivers[0].Name).To(Equal("SomeDriver"))
		})
	})

})

type errCloser struct{ io.Reader }

func (errCloser) Close() error                     { return nil }
func (errCloser) Read(p []byte) (n int, err error) { return 0, fmt.Errorf("any") }

type stringCloser struct{ io.Reader }

func (stringCloser) Close() error { return nil }
