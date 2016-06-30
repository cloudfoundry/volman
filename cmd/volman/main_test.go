package main_test

import (
	"fmt"
	"os"
	"os/exec"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/cloudfoundry-incubator/voldriver"
	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/volhttp"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Volman main", func() {
	var (
		args        []string
		listenAddr  string
		driversPath string
		process     ifrit.Process
		client      volman.Manager
		testLogger  lager.Logger
	)

	BeforeEach(func() {
		driversPath = fmt.Sprintf("/tmp/testdrivers/%d/", GinkgoParallelNode())
		err := os.MkdirAll(driversPath, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		listenAddr = fmt.Sprintf("0.0.0.0:%d", 8889+GinkgoParallelNode())
		client = volhttp.NewRemoteClient("http://" + listenAddr)
		Expect(err).NotTo(HaveOccurred())

		testLogger = lagertest.NewTestLogger("test")
	})

	JustBeforeEach(func() {
		args = append(args, "--listenAddr", listenAddr)
		args = append(args, "--volmanDriverPaths", driversPath)

		volmanRunner := ginkgomon.New(ginkgomon.Config{
			Name:       "volman",
			Command:    exec.Command(binaryPath, args...),
			StartCheck: "started",
		})
		process = ginkgomon.Invoke(volmanRunner)
	})

	AfterEach(func() {
		ginkgomon.Kill(process)
		err := os.RemoveAll(driversPath)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should listen on the given address", func() {
		_, err := client.ListDrivers(testLogger)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("given a driverspath with a single spec file", func() {
		BeforeEach(func() {
			err := voldriver.WriteDriverSpec(testLogger, driversPath, "test-driver", "spec", []byte("http://doesnotdoanything"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("should look in that location for driver specs", func() {
			response, err := client.ListDrivers(testLogger)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(response.Drivers)).To(Equal(1))
		})
	})
})
