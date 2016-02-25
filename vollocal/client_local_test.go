package vollocal_test

import (
	"fmt"

	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/vollocal"
	"github.com/cloudfoundry-incubator/volman/volmanfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("Volman", func() {

	Context("has no drivers", func() {
		BeforeEach(func() {
			client = &vollocal.LocalClient{vollocal.NewDriverClientCli("", nil)}
		})

		It("should report empty list of drivers", func() {
			testLogger := lagertest.NewTestLogger("ClientTest")
			drivers, err := client.ListDrivers(testLogger)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(drivers.Drivers)).To(Equal(0))
		})
	})

	Context("has drivers", func() {
		var fakeDriverClient *volmanfakes.FakeDriverPlugin

		BeforeEach(func() {
			fakeDriverClient = new(volmanfakes.FakeDriverPlugin)
			client = &vollocal.LocalClient{fakeDriverClient}
		})

		It("should fail to list if an error occurs listing", func() {

			fakeDriverClient.ListDriversReturns(nil, fmt.Errorf("Error Listing drivers"))

			_, err := whenListDriversIsRan()

			Expect(err).To(HaveOccurred())
		})

		It("should report list of drivers", func() {

			driver := volman.Driver{
				Name: "SomeDriver",
			}
			driverList := []volman.Driver{driver}
			fakeDriverClient.ListDriversReturns(driverList, nil)

			drivers, err := whenListDriversIsRan()
			Expect(err).NotTo(HaveOccurred())
			Expect(drivers.Drivers[0].Name).To(Equal("SomeDriver"))
		})

		It("should mount a volume, given a driver name, version, volume id, and opaque blob of configuration", func() {

			testLogger := lagertest.NewTestLogger("ClientTest")

			driverId := "SomeDriver"
			volumeId := "fake-volume"
			config := "Here is some config!"

			fakeDriverClient.MountReturns("/mnt/something", nil)

			mountPoint, err := client.Mount(testLogger, driverId, volumeId, config)

			Expect(err).NotTo(HaveOccurred())
			Expect(mountPoint.Path).To(Equal("/mnt/something"))
		})

		It("should fail if underlying mount command fails", func() {
			fakeDriverClient.MountReturns("", fmt.Errorf("any"))
			testLogger := lagertest.NewTestLogger("ClientTest")

			_, err := client.Mount(testLogger, "SomeDriver", "volumeId", "config")

			Expect(err).To(HaveOccurred())
		})
	})

})

func whenListDriversIsRan() (volman.ListDriversResponse, error) {
	testLogger := lagertest.NewTestLogger("ClientTest")
	return client.ListDrivers(testLogger)
}
