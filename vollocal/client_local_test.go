package vollocal_test

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/cloudfoundry-incubator/volman/vollocal"
	"github.com/cloudfoundry-incubator/volman/volmanfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("Volman", func() {

	var fakeClientFactory *volmanfakes.FakeRemoteClientFactory
	var fakeClient *volmanfakes.FakeDriver
	var validDriverInfoResponse io.ReadCloser
	var testLogger = lagertest.NewTestLogger("ClientTest")

	var driverName string

	BeforeEach(func() {
		driverName = "fakedriver"

		validDriverInfoResponse = stringCloser{bytes.NewBufferString("{\"Name\":\"fakedriver\",\"Path\":\"somePath\"}")}
	})

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
			writeDriverSpec(driverName, fmt.Sprintf("{\"Name\": \"fakedriver\",\"Addr\": \"http://0.0.0.0:%d\"}", 8080))
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

	Context("when given valid driver path", func() {

		BeforeEach(func() {
			fakeClientFactory = new(volmanfakes.FakeRemoteClientFactory)
			fakeClient = new(volmanfakes.FakeDriver)
			fakeClientFactory.NewRemoteClientReturns(fakeClient, nil)

			writeDriverSpec(driverName, fmt.Sprintf("{\"Name\": \"fakedriver\",\"Addr\": \"http://0.0.0.0:%d\"}", 8080))
			client = vollocal.NewLocalClientWithRemoteClientFactory(defaultPluginsDirectory, fakeClientFactory)
			driverName = "fakedriver"
		})

		It("should be able to mount", func() {

			volumeId := "fake-volume"
			config := "Here is some config!"

			mountPath, err := client.Mount(testLogger, driverName, volumeId, config)
			Expect(err).NotTo(HaveOccurred())
			Expect(mountPath).NotTo(Equal(""))
		})

		It("should not be able to mount if mount fails", func() {
			mountResponse := voldriver.MountResponse{}
			fakeClient.MountReturns(mountResponse, fmt.Errorf("an error"))

			volumeId := "fake-volume"
			config := "Here is some config!"

			_, err := client.Mount(testLogger, driverName, volumeId, config)
			Expect(err).To(HaveOccurred())
		})

		It("should be able to unmount", func() {
			volumeId := "fake-volume"

			err := client.Unmount(testLogger, driverName, volumeId)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when there is a malformed json driver spec", func() {

			BeforeEach(func() {
				driverName = "invalid-driver"

				writeDriverSpec(driverName, "invalid json")
			})

			It("should not be able to mount", func() {

				volumeId := "fake-volume"
				config := "Here is some config!"
				_, err := client.Mount(testLogger, driverName, volumeId, config)
				Expect(err).To(HaveOccurred())
			})

			It("should not be able to unmount", func() {
				volumeId := "fake-volume"

				err := client.Unmount(testLogger, driverName, volumeId)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when given invalid driver", func() {

			BeforeEach(func() {
				driverName = "does-not-exist"
			})

			It("should not be able to mount", func() {
				volumeId := "fake-volume"
				config := "Here is some config!"
				_, err := client.Mount(testLogger, driverName, volumeId, config)
				Expect(err).To(HaveOccurred())
			})

			It("should not be able to unmount", func() {
				volumeId := "fake-volume"

				err := client.Unmount(testLogger, driverName, volumeId)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})

func whenListDriversIsRan() (volman.ListDriversResponse, error) {
	testLogger := lagertest.NewTestLogger("ClientTest")
	return client.ListDrivers(testLogger)
}

func writeDriverSpec(driver string, contents string) {
	f, err := os.Create(defaultPluginsDirectory + "/" + driver + ".json")
	if err != nil {
		fmt.Printf("Error creating file " + err.Error())
	}
	defer f.Close()
	_, err = f.WriteString(contents)
	if err != nil {
		fmt.Printf("Error writing to file " + err.Error())
	}
	f.Sync()

}
