package main_test

import (
	"fmt"
	"time"

	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/cloudfoundry-incubator/volman/voldriver/driverhttp"

	"os"
	"path/filepath"

	"github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("UNIX and TCP versions of the fakedriver", func() {

	FakeDriverBehaviorWhenTransportIs("unix", func() (voldriver.Driver, ifrit.Runner) {
		return driverhttp.NewRemoteUnixClient(socketPath), unixRunner
	})

	FakeDriverBehaviorWhenTransportIs("tcp", func() (voldriver.Driver, ifrit.Runner) {
		return driverhttp.NewRemoteClient(fmt.Sprintf("http://0.0.0.0:%d", fakedriverServerPort)), runner
	})

})

var FakeDriverBehaviorWhenTransportIs = func(described string, args func() (voldriver.Driver, ifrit.Runner)) {

	Describe("Fake Driver Integration", func() {

		var testLogger = lagertest.NewTestLogger("FakeDriver Server Test")
		var client voldriver.Driver
		var ourRunner ifrit.Runner

		Context(fmt.Sprintf("given a started fakedriver %s server", described), func() {

			var (
				volumeId   string
				volumeName string
				opts       map[string]interface{}
			)

			BeforeEach(func() {
				uuid, err := uuid.NewV4()
				Expect(err).NotTo(HaveOccurred())
				volumeId = "fake-volume-id_" + uuid.String()
				volumeName = "fake-volume-name_" + uuid.String()
				opts = map[string]interface{}{"volume_id": volumeId}

				client, ourRunner = args()

				fakedriverServerProcess = ginkgomon.Invoke(ourRunner)
				time.Sleep(time.Millisecond * 1000)
			})

			It("should be able to create a volume", func() {
				createVolume(testLogger, client, volumeName, opts)
			})

			Context("given a created volume", func() {

				var mountResponse voldriver.MountResponse

				JustBeforeEach(func() {
					createVolume(testLogger, client, volumeName, opts)
				})

				It("should be able to mount a volume", func() {
					mountResponse = mountVolume(testLogger, client, volumeName)
				})

				It("should be able to remove the volume", func() {
					removeVolume(testLogger, client, volumeName)
				})

				Context("given a mounted volume", func() {

					JustBeforeEach(func() {
						mountResponse = mountVolume(testLogger, client, volumeName)
					})

					It("should exist", func() {
						Expect(mountResponse.Mountpoint).NotTo(Equal(""))
						defer os.Remove(mountResponse.Mountpoint)

						numberOfExistingMountPoints(mountResponse, 1)
					})

					It("should unmount a volume given same volume ID", func() {
						unmountRequest := voldriver.UnmountRequest{Name: volumeName}
						unmountErr := client.Unmount(testLogger, unmountRequest)
						Expect(unmountErr.Err).To(Equal(""))

						numberOfExistingMountPoints(mountResponse, 0)
					})

					It("should be able to remove the volume", func() {
						removeVolume(testLogger, client, volumeName)
						numberOfExistingMountPoints(mountResponse, 0)
					})
				})
			})
		})
	})
}

func createVolume(testLogger *lagertest.TestLogger, client voldriver.Driver, volumeName string, opts map[string]interface{}) {
	testLogger.Info("creating-volume", lager.Data{"name": volumeName})

	createRequest := voldriver.CreateRequest{Name: volumeName, Opts: opts}
	createResponse := client.Create(testLogger, createRequest)

	Expect(createResponse.Err).To(Equal(""))
}

func mountVolume(testLogger *lagertest.TestLogger, client voldriver.Driver, volumeName string) voldriver.MountResponse {
	testLogger.Info("mounting-volume", lager.Data{"name": volumeName})

	mountRequest := voldriver.MountRequest{Name: volumeName}
	mountResponse := client.Mount(testLogger, mountRequest)

	Expect(mountResponse.Err).To(Equal(""))

	return mountResponse
}

func removeVolume(testLogger *lagertest.TestLogger, client voldriver.Driver, volumeName string) {
	removeRequest := voldriver.RemoveRequest{Name: volumeName}
	removeErr := client.Remove(testLogger, removeRequest)
	Expect(removeErr.Err).To(Equal(""))

	getRequest := voldriver.GetRequest{Name: volumeName}
	getErr := client.Get(testLogger, getRequest)
	Expect(getErr.Err).To(Equal("Volume not found"))
}

func numberOfExistingMountPoints(mountResponse voldriver.MountResponse, num int) {
	matches, err := filepath.Glob(mountResponse.Mountpoint)
	Expect(err).NotTo(HaveOccurred())
	Expect(len(matches)).To(Equal(num))
}
