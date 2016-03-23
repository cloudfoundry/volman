package vollocal_test

import (
	"bytes"
	"io"
	"time"

	"fmt"

	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/cloudfoundry-incubator/volman/vollocal"
	"github.com/cloudfoundry-incubator/volman/volmanfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Volman", func() {
	var fakeDriverFactory *volmanfakes.FakeDriverFactory
	var fakeDriver *volmanfakes.FakeDriver
	var validDriverInfoResponse io.ReadCloser
	var testLogger = lagertest.NewTestLogger("ClientTest")

	var driverName string

	BeforeEach(func() {
		driverName = "fakedriver"
		validDriverInfoResponse = stringCloser{bytes.NewBufferString("{\"Name\":\"fakedriver\",\"Path\":\"somePath\"}")}
	})

	Describe("ListDrivers", func() {
		Context("has no drivers in location", func() {
			BeforeEach(func() {
				client = vollocal.NewLocalClient("/noplugins")
			})

			It("should report empty list of drivers", func() {
				drivers, err := client.ListDrivers(testLogger)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(drivers.Drivers)).To(Equal(0))
			})
		})
		Context("when location is not set", func() {
			BeforeEach(func() {
				client = vollocal.NewLocalClient("")
			})

			It("should report empty list of drivers", func() {
				drivers, err := client.ListDrivers(testLogger)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(drivers.Drivers)).To(Equal(0))
			})
		})
		Context("has driver in location", func() {
			BeforeEach(func() {
				err := voldriver.WriteDriverSpec(testLogger, defaultPluginsDirectory, driverName, "http://0.0.0.0:8080")
				Expect(err).NotTo(HaveOccurred())

				client = vollocal.NewLocalClient(defaultPluginsDirectory)
			})

			It("should report list of drivers", func() {
				drivers, err := client.ListDrivers(testLogger)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(drivers.Drivers)).ToNot(Equal(0))
			})

			It("should report at least fakedriver", func() {
				drivers, err := client.ListDrivers(testLogger)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(drivers.Drivers)).ToNot(Equal(0))
				Expect(drivers.Drivers[0].Name).To(Equal("fakedriver"))
			})
		})
		Context("discovery fails", func() {
			It("it should fail", func() {
				fakeDriverFactory = new(volmanfakes.FakeDriverFactory)

				fakeDriverFactory.DiscoverReturns(nil, fmt.Errorf("error discovering drivers"))
				driverName = "fakedriver"

				client = vollocal.NewLocalClientWithDriverFactory(fakeDriverFactory)
				_, err := client.ListDrivers(testLogger)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Mount and Unmount", func() {
		Context("when given valid driver", func() {
			BeforeEach(func() {
				fakedriverProcess = ginkgomon.Invoke(fakedriverRunner)
				time.Sleep(time.Millisecond * 1000)

				fakeDriverFactory = new(volmanfakes.FakeDriverFactory)
				fakeDriver = new(volmanfakes.FakeDriver)

				fakeDriverFactory.DriverReturns(fakeDriver, nil)
				driverName = "fakedriver"

				client = vollocal.NewLocalClientWithDriverFactory(fakeDriverFactory)
			})

			It("should be able to mount", func() {
				volumeId := "fake-volume"

				mountPath, err := client.Mount(testLogger, driverName, volumeId, map[string]interface{}{"volume_id": volumeId})
				Expect(err).NotTo(HaveOccurred())
				Expect(mountPath).NotTo(Equal(""))
			})

			It("should not be able to mount if mount fails", func() {
				mountResponse := voldriver.MountResponse{Err: "an error"}
				fakeDriver.MountReturns(mountResponse)

				volumeId := "fake-volume"
				_, err := client.Mount(testLogger, driverName, volumeId, map[string]interface{}{"volume_id": volumeId})
				Expect(err).To(HaveOccurred())
			})

			It("should be able to unmount", func() {
				volumeId := "fake-volume"

				err := client.Unmount(testLogger, driverName, volumeId)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should not be able to unmount when driver unmount fails", func() {
				fakeDriver.UnmountReturns(voldriver.ErrorResponse{Err: "unmount failure"})
				volumeId := "fake-volume"

				err := client.Unmount(testLogger, driverName, volumeId)
				Expect(err).To(HaveOccurred())
			})

		})
		Context("when given invalid driver", func() {
			BeforeEach(func() {
				fakedriverProcess = ginkgomon.Invoke(fakedriverRunner)
				time.Sleep(time.Millisecond * 1000)

				fakeDriverFactory = new(volmanfakes.FakeDriverFactory)
				fakeDriver = new(volmanfakes.FakeDriver)

				fakeDriverFactory.DriverReturns(fakeDriver, nil)
				driverName = "fakedriver"

				client = vollocal.NewLocalClientWithDriverFactory(fakeDriverFactory)
				fakeDriverFactory.DriverReturns(nil, fmt.Errorf("driver not found"))
			})

			It("should not be able to mount", func() {
				volumeId := "fake-volume"
				_, err := client.Mount(testLogger, driverName, volumeId, map[string]interface{}{"volume_id": volumeId})
				Expect(err).To(HaveOccurred())
			})

			It("should not be able to unmount", func() {
				volumeId := "fake-volume"

				err := client.Unmount(testLogger, driverName, volumeId)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("after creating successfully driver is not found", func() {
			BeforeEach(func() {
				fakedriverProcess = ginkgomon.Invoke(fakedriverRunner)
				time.Sleep(time.Millisecond * 1000)

				fakeDriverFactory = new(volmanfakes.FakeDriverFactory)
				fakeDriver = new(volmanfakes.FakeDriver)

				fakeDriverFactory.DriverReturns(fakeDriver, nil)
				driverName = "fakedriver"

				client = vollocal.NewLocalClientWithDriverFactory(fakeDriverFactory)
				calls := 0
				fakeDriverFactory.DriverStub = func(lager.Logger, string) (voldriver.Driver, error) {
					calls++
					if calls > 1 {
						return nil, fmt.Errorf("driver not found")
					}
					return fakeDriver, nil
				}
			})
			It("should not be able to mount", func() {
				volumeId := "fake-volume"
				_, err := client.Mount(testLogger, driverName, volumeId, map[string]interface{}{"volume_id": volumeId})
				Expect(err).To(HaveOccurred())
			})

		})
		Context("after unsuccessfully creating", func() {
			BeforeEach(func() {
				fakedriverProcess = ginkgomon.Invoke(fakedriverRunner)
				time.Sleep(time.Millisecond * 1000)

				fakeDriverFactory = new(volmanfakes.FakeDriverFactory)
				fakeDriver = new(volmanfakes.FakeDriver)

				driverName = "fakedriver"
				client = vollocal.NewLocalClientWithDriverFactory(fakeDriverFactory)
				fakeDriverFactory.DriverReturns(fakeDriver, nil)

				fakeDriver.CreateReturns(voldriver.ErrorResponse{"create fails"})
			})
			It("should not be able to mount", func() {
				volumeId := "fake-volume"
				_, err := client.Mount(testLogger, driverName, volumeId, map[string]interface{}{"volume_id": volumeId})
				Expect(err).To(HaveOccurred())
			})

		})
	})

})

func whenListDriversIsRan() (volman.ListDriversResponse, error) {
	testLogger := lagertest.NewTestLogger("ClientTest")
	return client.ListDrivers(testLogger)
}
