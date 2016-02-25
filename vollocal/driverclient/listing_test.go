package driverclient_test

import (
	"bytes"
	"fmt"

	"github.com/cloudfoundry-incubator/volman/vollocal/driverclient"
	"github.com/cloudfoundry-incubator/volman/volmanfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("Listing", func() {
	var subject *driverclient.Listing
	var fakeCmd *volmanfakes.FakeCmd
	var fakeExec *volmanfakes.FakeExec
	var testLogger = lagertest.NewTestLogger("ClientTest")

	BeforeEach(func() {
		fakeExec = new(volmanfakes.FakeExec)
		fakeCmd = new(volmanfakes.FakeCmd)
		fakeExec.CommandReturns(fakeCmd)
	})

	Context("when no drivers are present", func() {
		BeforeEach(func() {
			subject = &driverclient.Listing{[]string{}, fakeExec}
		})

		It("should not report an error or list any drivers", func() {
			volmanDrivers, err := subject.List(testLogger, "/path/to/drivers")
			Expect(err).ToNot(HaveOccurred())
			Expect(len(volmanDrivers)).To(Equal(0))
		})
	})

	Context("when drivers are present", func() {
		BeforeEach(func() {
			list := []string{"binary1"}
			subject = &driverclient.Listing{list, fakeExec}
		})

		It("should list drivers and not report an error", func() {
			fakeCmd.StdoutPipeReturns(stringCloser{bytes.NewBufferString("{\"Name\":\"binary1\"}")}, nil)
			volmanDrivers, err := subject.List(testLogger, "/path/to/drivers")
			Expect(err).ToNot(HaveOccurred())
			Expect(len(volmanDrivers)).To(Equal(1))
			Expect(volmanDrivers[0].Name).To(Equal("binary1"))
			executable, args := fakeExec.CommandArgsForCall(0)
			Expect(executable).To(Equal("binary1"))
			Expect(args[0]).To(Equal("info"))
		})

		It("should report an error if binary fails", func() {
			fakeCmd.StdoutPipeReturns(stringCloser{bytes.NewBufferString("{\"Name\":\"binary1\"}")}, nil)
			fakeCmd.WaitReturns(fmt.Errorf("binary fails"))
			_, err := subject.List(testLogger, "/path/to/drivers")
			Expect(err).To(HaveOccurred())
		})
	})
})
