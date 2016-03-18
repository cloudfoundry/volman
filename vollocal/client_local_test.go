package vollocal_test

import (
	"bytes"
	"io"
	"time"

	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/cloudfoundry-incubator/volman/vollocal"
	"github.com/cloudfoundry-incubator/volman/volmanfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Volman", func() {
	var fakeClientFactory *volmanfakes.FakeRemoteClientFactory
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
	})

	Describe("Mount and Unmount with TCP config", func() {
		var tcpclient volman.Manager
		var err error
		Context("when given valid driver path", func() {
			BeforeEach(func() {
				fakedriverProcess = ginkgomon.Invoke(fakedriverRunner)
				time.Sleep(time.Millisecond * 1000)

				fakeClientFactory = new(volmanfakes.FakeRemoteClientFactory)

				fakeDriver = new(volmanfakes.FakeDriver)
				fakeClientFactory.NewRemoteClientReturns(fakeDriver, nil)
				driverName = "fakedriver"
				err := voldriver.WriteDriverSpec(testLogger, defaultPluginsDirectory, driverName, "http://0.0.0.0:8080")
				Expect(err).NotTo(HaveOccurred())
				tcpclient = vollocal.NewLocalClientWithRemoteClientFactory(defaultPluginsDirectory, fakeClientFactory)
			})

			It("should be able to mount", func() {
				volumeId := "fake-volume"

				mountPath, err := tcpclient.Mount(testLogger, driverName, volumeId, map[string]interface{}{"volume_id": volumeId})
				Expect(err).NotTo(HaveOccurred())
				Expect(mountPath).NotTo(Equal(""))
			})

			It("should not be able to mount if mount fails", func() {
				mountResponse := voldriver.MountResponse{Err: "an error"}
				fakeDriver.MountReturns(mountResponse)

				volumeId := "fake-volume"
				_, err = tcpclient.Mount(testLogger, driverName, volumeId, map[string]interface{}{"volume_id": volumeId})
				Expect(err).To(HaveOccurred())
			})

			It("should be able to unmount", func() {
				volumeId := "fake-volume"

				err = tcpclient.Unmount(testLogger, driverName, volumeId)
				Expect(err).NotTo(HaveOccurred())
			})
			Context("when given invalid driver", func() {

				BeforeEach(func() {
					driverName = "does-not-exist"
				})

				It("should not be able to mount", func() {
					volumeId := "fake-volume"
					_, err = tcpclient.Mount(testLogger, driverName, volumeId, map[string]interface{}{"volume_id": volumeId})
					Expect(err).To(HaveOccurred())
				})

				It("should not be able to unmount", func() {
					volumeId := "fake-volume"

					err = tcpclient.Unmount(testLogger, driverName, volumeId)
					Expect(err).To(HaveOccurred())
				})
			})
		})

	})
	Describe("Mount and Unmount with unix config", func() {
		var unixclient volman.Manager
		var err error
		Context("when given valid driver path", func() {
			BeforeEach(func() {
				fakedriverUnixServerProcess = ginkgomon.Invoke(unixRunner)
				time.Sleep(time.Millisecond * 1000)
				fakeClientFactory = new(volmanfakes.FakeRemoteClientFactory)

				fakeDriver = new(volmanfakes.FakeDriver)
				fakeClientFactory.NewRemoteClientReturns(fakeDriver, nil)
				driverName = "fakedriver"
				Expect(err).NotTo(HaveOccurred())
				unixclient = vollocal.NewLocalClientWithRemoteClientFactory(defaultPluginsDirectory, fakeClientFactory)
				driverName = "fakedriver"
			})

			It("should be able to mount", func() {
				volumeId := "fake-volume"

				mountPath, err := unixclient.Mount(testLogger, driverName, volumeId, map[string]interface{}{"volume_id": volumeId})
				Expect(err).NotTo(HaveOccurred())
				Expect(mountPath).NotTo(Equal(""))
			})

			It("should not be able to mount if mount fails", func() {
				mountResponse := voldriver.MountResponse{Err: "an error"}
				fakeDriver.MountReturns(mountResponse)

				volumeId := "fake-volume"
				_, err = unixclient.Mount(testLogger, driverName, volumeId, map[string]interface{}{"volume_id": volumeId})
				Expect(err).To(HaveOccurred())
			})

			It("should be able to unmount", func() {
				volumeId := "fake-volume"

				err = unixclient.Unmount(testLogger, driverName, volumeId)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when given invalid driver", func() {

				BeforeEach(func() {
					driverName = "does-not-exist"
				})

				It("should not be able to mount", func() {
					volumeId := "fake-volume"
					_, err = unixclient.Mount(testLogger, driverName, volumeId, map[string]interface{}{"volume_id": volumeId})
					Expect(err).To(HaveOccurred())
				})

				It("should not be able to unmount", func() {
					volumeId := "fake-volume"

					err = unixclient.Unmount(testLogger, driverName, volumeId)
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})

func whenListDriversIsRan() (volman.ListDriversResponse, error) {
	testLogger := lagertest.NewTestLogger("ClientTest")
	return client.ListDrivers(testLogger)
}
