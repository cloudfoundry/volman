package driverclient_test

import (
	"bytes"
	"fmt"

	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/vollocal/driverclient"
	"github.com/cloudfoundry-incubator/volman/volmanfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("Mounting", func() {
	var subject *driverclient.Mounting
	var fakeCmd *volmanfakes.FakeCmd
	var fakeExec *volmanfakes.FakeExec
	var Driver *volman.Driver

	BeforeEach(func() {
		fakeExec = new(volmanfakes.FakeExec)
		fakeCmd = new(volmanfakes.FakeCmd)
		fakeExec.CommandReturns(fakeCmd)

		Driver = &volman.Driver{
			Name: "SomeDriver",
		}
	})

	Context("when no binary is found", func() {
		BeforeEach(func() {
			subject = &driverclient.Mounting{*Driver, nil, fakeExec}
		})

		It("should report an error", func() {
			testLogger := lagertest.NewTestLogger("ClientTest")
			_, err := subject.Mount(testLogger, "", "")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when binary is found", func() {
		BeforeEach(func() {
			list := []volman.Driver{volman.Driver{
				Name:   "SomeDriver",
				Binary: "/somebin",
			}}
			subject = &driverclient.Mounting{*Driver, list, fakeExec}
		})

		It("should mount", func() {
			fakeCmd.StdoutPipeReturns(stringCloser{bytes.NewBufferString("{\"Path\":\"/mnt/something\"}")}, nil)
			testLogger := lagertest.NewTestLogger("ClientTest")
			_, err := subject.Mount(testLogger, "", "")
			Expect(err).ToNot(HaveOccurred())
			executable, args := fakeExec.CommandArgsForCall(0)
			Expect(executable).To(Equal("/somebin"))
			Expect(args[0]).To(Equal("mount"))
		})
		Context("when binary but fails to mount", func() {
			It("should report an error", func() {
				fakeCmd.StdoutPipeReturns(stringCloser{bytes.NewBufferString("")}, fmt.Errorf("any"))
				testLogger := lagertest.NewTestLogger("ClientTest")
				_, err := subject.Mount(testLogger, "", "")
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
