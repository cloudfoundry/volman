package vollocal_test

import (
	"bytes"
	"io"
	"time"

	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/cloudfoundry-incubator/volman/vollocal"
	"github.com/cloudfoundry-incubator/volman/volmanfakes"

	"github.com/pivotal-golang/clock/fakeclock"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Volman", func() {
	var (
		logger = lagertest.NewTestLogger("client-test")

		fakeDriverFactory *volmanfakes.FakeDriverFactory
		fakeDriver        *volmanfakes.FakeDriver
		fakeClock         *fakeclock.FakeClock

		scanInterval time.Duration

		validDriverInfoResponse io.ReadCloser

		driverSyncer vollocal.DriverSyncer

		runner  ifrit.Runner
		process ifrit.Process
	)

	BeforeEach(func() {
		fakeDriverFactory = new(volmanfakes.FakeDriverFactory)
		fakeClock = fakeclock.NewFakeClock(time.Unix(123, 456))

		scanInterval = 1 * time.Second

		validDriverInfoResponse = stringCloser{bytes.NewBufferString("{\"Name\":\"fakedriver\",\"Path\":\"somePath\"}")}
	})

	Describe("ListDrivers", func() {

		BeforeEach(func() {
			driverSyncer = vollocal.NewDriverSyncer(logger, fakeDriverFactory, scanInterval, fakeClock)
			client, runner = vollocal.NewLocalClientWithDriverFactory(logger, defaultPluginsDirectory, fakeDriverFactory, driverSyncer)
			process = ginkgomon.Invoke(runner)
		})

		It("should report empty list of drivers", func() {
			drivers, err := client.ListDrivers(logger)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(drivers.Drivers)).To(Equal(0))
		})

		Context("has no drivers in location", func() {

			BeforeEach(func() {
				fakeDriverFactory = new(volmanfakes.FakeDriverFactory)
				fakeDriverFactory.DriversDirReturns("")
			})

			It("should report empty list of drivers", func() {
				drivers, err := client.ListDrivers(logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(drivers.Drivers)).To(Equal(0))
			})

			AfterEach(func() {
				ginkgomon.Kill(process)
			})

		})

		Context("has driver in location", func() {
			BeforeEach(func() {
				err := voldriver.WriteDriverSpec(logger, defaultPluginsDirectory, "fakedriver", "spec",[]byte("http://0.0.0.0:8080"))
				Expect(err).NotTo(HaveOccurred())

				driverSyncer = vollocal.NewDriverSyncer(logger, fakeDriverFactory, scanInterval, fakeClock)
				client, runner = vollocal.NewLocalClientWithDriverFactory(logger, defaultPluginsDirectory, fakeDriverFactory, driverSyncer)
			})

			JustBeforeEach(func() {
				process = ginkgomon.Invoke(runner)
			})

			AfterEach(func() {
				ginkgomon.Kill(process)
			})

			It("should report empty list of drivers", func() {
				drivers, err := client.ListDrivers(logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(drivers.Drivers)).To(Equal(0))
			})

			Context("after running drivers discovery", func() {

				BeforeEach(func() {
					fakeDriver := new(volmanfakes.FakeDriver)
					fakeDriverFactory.DiscoverReturns(map[string]voldriver.Driver{"fakedriver": fakeDriver}, nil)
					fakeDriverFactory.DriverReturns(fakeDriver, nil)

				})

				It("should report at least fakedriver", func() {
					drivers, err := client.ListDrivers(logger)
					Expect(err).NotTo(HaveOccurred())
					Expect(len(drivers.Drivers)).ToNot(Equal(0))
					Expect(drivers.Drivers[0].Name).To(Equal("fakedriver"))
				})

				It("all reported drivers should not be activated", func() {
					drivers, err := client.ListDrivers(logger)
					Expect(err).NotTo(HaveOccurred())

					for _, driver := range drivers.Drivers {
						activated, err := driverSyncer.DriverRegistry().Activated(driver.Name)
						Expect(err).NotTo(HaveOccurred())
						Expect(activated).To(BeFalse())
					}
				})
			})
		})

		Context("discovery fails", func() {
			It("it should fail", func() {
				//TODO
			})
		})
	})

	Describe("Mount and Unmount", func() {
		Context("when given valid driver", func() {
			BeforeEach(func() {
				fakeDriverFactory = new(volmanfakes.FakeDriverFactory)
				fakeDriver = new(volmanfakes.FakeDriver)
				fakeDriverFactory.DriverReturns(fakeDriver, nil)

				drivers := make(map[string]voldriver.Driver)
				drivers["fakedriver"] = fakeDriver

				fakeDriverFactory.DiscoverReturns(drivers, nil)

				err := voldriver.WriteDriverSpec(logger, defaultPluginsDirectory, "fakedriver", "spec", []byte(fmt.Sprintf("http://0.0.0.0:%d", localDriverServerPort)))
				Expect(err).NotTo(HaveOccurred())

				driverSyncer = vollocal.NewDriverSyncer(logger, fakeDriverFactory, scanInterval, fakeClock)
				client, runner = vollocal.NewLocalClientWithDriverFactory(logger, defaultPluginsDirectory, fakeDriverFactory, driverSyncer)

				process = ginkgomon.Invoke(runner)
			})

			AfterEach(func() {
				ginkgomon.Kill(process)
			})

			Context("mount", func() {
				It("should be able to mount and activate", func() {
					volumeId := "fake-volume"

					mountPath, err := client.Mount(logger, "fakedriver", volumeId, map[string]interface{}{"volume_id": volumeId})
					Expect(err).NotTo(HaveOccurred())
					Expect(mountPath).NotTo(Equal(""))

					activated, err := driverSyncer.DriverRegistry().Activated("fakedriver")
					Expect(err).NotTo(HaveOccurred())
					Expect(activated).To(BeTrue())
				})

				It("should not be able to mount if mount fails", func() {
					mountResponse := voldriver.MountResponse{Err: "an error"}
					fakeDriver.MountReturns(mountResponse)

					volumeId := "fake-volume"
					_, err := client.Mount(logger, "fakedriver", volumeId, map[string]interface{}{"volume_id": volumeId})
					Expect(err).To(HaveOccurred())
				})
			})

			Context("umount", func() {
				It("should be able to unmount and activate", func() {
					volumeId := "fake-volume"

					err := client.Unmount(logger, "fakedriver", volumeId)
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeDriver.UnmountCallCount()).To(Equal(1))
					Expect(fakeDriver.RemoveCallCount()).To(Equal(1))

					activated, err := driverSyncer.DriverRegistry().Activated("fakedriver")
					Expect(err).NotTo(HaveOccurred())
					Expect(activated).To(BeTrue())
				})

				It("should not be able to unmount when driver unmount fails", func() {
					fakeDriver.UnmountReturns(voldriver.ErrorResponse{Err: "unmount failure"})
					volumeId := "fake-volume"

					err := client.Unmount(logger, "fakedriver", volumeId)
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Context("when given invalid driver", func() {
			BeforeEach(func() {
				localDriverProcess = ginkgomon.Invoke(localDriverRunner)
				fakeDriverFactory = new(volmanfakes.FakeDriverFactory)
				fakeDriver = new(volmanfakes.FakeDriver)

				fakeDriverFactory.DriverReturns(fakeDriver, nil)
				fakeDriverFactory.DriverReturns(nil, fmt.Errorf("driver not found"))

				driverSyncer = vollocal.NewDriverSyncer(logger, fakeDriverFactory, scanInterval, fakeClock)
				client, runner = vollocal.NewLocalClientWithDriverFactory(logger, defaultPluginsDirectory, fakeDriverFactory, driverSyncer)
				process = ginkgomon.Invoke(runner)
			})

			AfterEach(func() {
				ginkgomon.Kill(process)
			})

			It("should not be able to mount", func() {
				_, err := client.Mount(logger, "fakedriver", "fake-volume", map[string]interface{}{"volume_id": "fake-volume"})
				Expect(err).To(HaveOccurred())
			})

			It("should not be able to unmount", func() {
				err := client.Unmount(logger, "fakedriver", "fake-volume")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("after creating successfully driver is not found", func() {
			BeforeEach(func() {

				fakeDriverFactory = new(volmanfakes.FakeDriverFactory)
				fakeDriver = new(volmanfakes.FakeDriver)
				mountReturn := voldriver.MountResponse{Err: "driver not found",
					Mountpoint: "",
				}
				fakeDriver.MountReturns(mountReturn)
				fakeDriverFactory.DriverReturns(fakeDriver, nil)

				driverSyncer = vollocal.NewDriverSyncer(logger, fakeDriverFactory, scanInterval, fakeClock)
				client, runner = vollocal.NewLocalClientWithDriverFactory(logger, defaultPluginsDirectory, fakeDriverFactory, driverSyncer)
				process = ginkgomon.Invoke(runner)

				calls := 0
				fakeDriverFactory.DriverStub = func(lager.Logger, string, string) (voldriver.Driver, error) {
					calls++
					if calls > 1 {
						return nil, fmt.Errorf("driver not found")
					}
					return fakeDriver, nil
				}
			})

			AfterEach(func() {
				ginkgomon.Kill(process)
			})

			It("should not be able to mount", func() {
				_, err := client.Mount(logger, "fakedriver", "fake-volume", map[string]interface{}{"volume_id": "fake-volume"})
				Expect(err).To(HaveOccurred())
			})

		})

		Context("after unsuccessfully creating", func() {
			BeforeEach(func() {
				localDriverProcess = ginkgomon.Invoke(localDriverRunner)
				fakeDriver = new(volmanfakes.FakeDriver)

				fakeDriverFactory = new(volmanfakes.FakeDriverFactory)
				fakeDriverFactory.DriverReturns(fakeDriver, nil)

				fakeDriver.CreateReturns(voldriver.ErrorResponse{"create fails"})

				driverSyncer = vollocal.NewDriverSyncer(logger, fakeDriverFactory, scanInterval, fakeClock)
				client, runner = vollocal.NewLocalClientWithDriverFactory(logger, defaultPluginsDirectory, fakeDriverFactory, driverSyncer)
				process = ginkgomon.Invoke(runner)
			})

			AfterEach(func() {
				ginkgomon.Kill(process)
			})

			It("should not be able to mount", func() {
				_, err := client.Mount(logger, "fakedriver", "fake-volume", map[string]interface{}{"volume_id": "fake-volume"})
				Expect(err).To(HaveOccurred())
			})

		})
	})
})
