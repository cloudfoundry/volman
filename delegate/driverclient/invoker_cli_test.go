package driverclient_test

import (
	"bytes"
	"fmt"
	"io"

	"github.com/cloudfoundry-incubator/volman/delegate/driverclient"
	"github.com/cloudfoundry-incubator/volman/volmanfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("Invoker CLI", func() {
	var subject *driverclient.CliInvoker
	var fakeCmd *volmanfakes.FakeCmd
	var fakeExec *volmanfakes.FakeExec
	var validJson io.ReadCloser
	var testLogger = lagertest.NewTestLogger("ClientTest")

	Context("when invoking a driver", func() {
		BeforeEach(func() {
			fakeExec = new(volmanfakes.FakeExec)
			fakeCmd = new(volmanfakes.FakeCmd)

			validJson = stringCloser{bytes.NewBufferString("{\"Name\":\"JSON\"}")}

			subject = &driverclient.CliInvoker{
				UseExec: fakeExec,
				UseCmd:  fakeCmd,
			}
		})

		It("should report an error when unable to attach to stdout", func() {
			fakeCmd.StdoutPipeReturns(errCloser{bytes.NewBufferString("")}, fmt.Errorf("unable to attach to stdout"))
			err := subject.InvokeDriver(testLogger, new(interface{}))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("unable to attach to stdout"))
		})

		It("should report an error when unable to start binary", func() {
			fakeCmd.StdoutPipeReturns(errCloser{bytes.NewBufferString("cmdfails")}, nil)
			fakeCmd.StartReturns(fmt.Errorf("unable to start binary"))
			err := subject.InvokeDriver(testLogger, new(interface{}))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("unable to start binary"))
		})

		It("should report an error when JSON encoding fails", func() {
			fakeCmd.StdoutPipeReturns(errCloser{bytes.NewBufferString("")}, nil)
			err := subject.InvokeDriver(testLogger, new(interface{}))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("any"))
		})
		Context("when valid JSON is returned during binary execution", func() {
			var data = struct {
				Name string `json:"Name"`
			}{}

			BeforeEach(func() {
				fakeCmd.StdoutPipeReturns(validJson, nil)
			})

			It("should report an error when executing the driver binary fails", func() {
				fakeCmd.WaitReturns(fmt.Errorf("executing driver binary fails"))

				err := subject.InvokeDriver(testLogger, &data)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("executing driver binary fails"))
			})
			It("should fill out upon successful cli invokation", func() {

				err := subject.InvokeDriver(testLogger, &data)
				Expect(err).ToNot(HaveOccurred())
			})
		})

	})

})
