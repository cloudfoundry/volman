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
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Fake Driver Integration", func() {
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
			var volumeId string
			var volumeName string
			var mountResponse voldriver.MountResponse

			JustBeforeEach(func() {
				client := driverhttp.NewRemoteClient(fmt.Sprintf("http://0.0.0.0:%d", fakedriverServerPort))
				testLogger := lagertest.NewTestLogger("FakeDriver Server Test")
				node := GinkgoParallelNode()
				volumeId = "fake-volume-id_" + strconv.Itoa(node)
				volumeName = "fake-volume-name_" + strconv.Itoa(node)
				testLogger.Info("creating-volume", lager.Data{"name": volumeName})
				createRequest := voldriver.CreateRequest{Name: volumeName, Opts: map[string]interface{}{"volume_id": volumeId}}
				createResponse := client.Create(testLogger, createRequest)
				Expect(createResponse.Err).To(Equal(""))

				mountRequest := voldriver.MountRequest{Name: volumeName}
				mountResponse = client.Mount(testLogger, mountRequest)
				Expect(mountResponse.Err).To(Equal(""))
			})

			It("should exist", func() {
				Expect(mountResponse.Mountpoint).NotTo(Equal(""))
				defer os.Remove(mountResponse.Mountpoint)

				matches, err := filepath.Glob(mountResponse.Mountpoint)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(matches)).To(Equal(1))
			})

			It("should unmount a volume given same volume ID", func() {
				client := driverhttp.NewRemoteClient(fmt.Sprintf("http://0.0.0.0:%d", fakedriverServerPort))
				testLogger := lagertest.NewTestLogger("FakeDriver Server Test")
				unmountRequest := voldriver.UnmountRequest{Name: volumeName}
				unmountErr := client.Unmount(testLogger, unmountRequest)
				Expect(unmountErr.Err).To(Equal(""))

				matches, err := filepath.Glob(mountResponse.Mountpoint)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(matches)).To(Equal(0))
			})

			AfterEach(func() {
				os.Remove(mountResponse.Mountpoint)
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
