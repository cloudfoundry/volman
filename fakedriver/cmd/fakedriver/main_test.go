package main_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/cloudfoundry-incubator/volman/voldriver/driverhttp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("FakeDriverServer", func() {

	BeforeEach(func() {
		fakedriverServerProcess = ginkgomon.Invoke(runner)
		time.Sleep(time.Millisecond * 1000)
	})

	Context("given a started fakedriver server", func() {

		It("should not exit", func() {
			Consistently(runner).ShouldNot(Exit())
		})

		It("should get a 404 for root", func() {
			_, status, err := get("/")
			Expect(err).NotTo(HaveOccurred())
			Expect(status).Should(ContainSubstring("404"))
		})

		It("should have a debug server endpoint", func() {
			_, err := http.Get(fmt.Sprintf("http://%s/debug/pprof/goroutine", debugServerAddress))
			Expect(err).NotTo(HaveOccurred())
		})

		Context("given a mounted volume", func() {
			var err error
			var volumeId string
			var mountPoint voldriver.MountResponse

			JustBeforeEach(func() {
				client := driverhttp.NewRemoteClient(fmt.Sprintf("http://0.0.0.0:%d", fakedriverServerPort))
				testLogger := lagertest.NewTestLogger("FakeDriver Server Test")
				node := GinkgoParallelNode()
				volumeId = "fake-volume_" + strconv.Itoa(node)
				testLogger.Info(fmt.Sprintf("Mounting volume: %s", volumeId))
				mountRequest := voldriver.MountRequest{VolumeId: volumeId}
				mountPoint, err = client.Mount(testLogger, mountRequest)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should exist", func() {
				Expect(mountPoint.Path).NotTo(Equal(""))
				defer os.Remove(mountPoint.Path)

				matches, err := filepath.Glob(mountPoint.Path)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(matches)).To(Equal(1))
			})

			It("should unmount a volume given same volume ID", func() {
				client := driverhttp.NewRemoteClient(fmt.Sprintf("http://0.0.0.0:%d", fakedriverServerPort))
				testLogger := lagertest.NewTestLogger("FakeDriver Server Test")
				unmountRequest := voldriver.UnmountRequest{VolumeId: volumeId}
				err := client.Unmount(testLogger, unmountRequest)
				Expect(err).NotTo(HaveOccurred())

				matches, err := filepath.Glob(mountPoint.Path)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(matches)).To(Equal(0))
			})

			AfterEach(func() {
				os.Remove(mountPoint.Path)
			})
		})
	})
})

func get(path string) (body string, status string, err error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://0.0.0.0:%d%s", fakedriverServerPort, path), nil)
	response, _ := (&http.Client{}).Do(req)
	defer response.Body.Close()
	bodyBytes, err := ioutil.ReadAll(response.Body)
	return string(bodyBytes[:]), response.Status, err
}
