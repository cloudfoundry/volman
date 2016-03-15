package main_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/cloudfoundry-incubator/volman/volhttp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Volman", func() {

	var (
		testLogger lager.Logger
	)

	BeforeEach(func() {
		testLogger = lagertest.NewTestLogger("MainTest")

		fakedriverProcess = ginkgomon.Invoke(fakedriverRunner)
		time.Sleep(time.Millisecond * 1000)

		volmanProcess = ginkgomon.Invoke(runner)
		time.Sleep(time.Millisecond * 1000)
	})

	Context("after starting", func() {
		It("should not exit", func() {
			Consistently(runner).ShouldNot(Exit())
		})
	})

	Context("after starting http server", func() {
		It("should get a 404 for root", func() {
			_, status, err := get("/")
			Expect(err).NotTo(HaveOccurred())
			Expect(status).Should(ContainSubstring("404"))
		})

		It("should return empty list for '/v1/drivers' (200 status)", func() {
			client := volhttp.NewRemoteClient(fmt.Sprintf("http://0.0.0.0:%d", volmanServerPort))
			drivers, err := client.ListDrivers(testLogger)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(drivers.Drivers)).To(Equal(0))
		})

		It("should have a debug server endpoint", func() {
			_, err := http.Get(fmt.Sprintf("http://%s/debug/pprof/goroutine", debugServerAddress))
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when driver installed in the spec file plugins directory", func() {
			BeforeEach(func() {
				err := voldriver.WriteDriverSpec(testLogger, tmpDriversPath, "fakedriver", fmt.Sprintf("http://0.0.0.0:%d", fakedriverServerPort))
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return list of drivers for '/v1/drivers' (200 status)", func() {
				client := volhttp.NewRemoteClient(fmt.Sprintf("http://0.0.0.0:%d", volmanServerPort))
				drivers, err := client.ListDrivers(testLogger)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(drivers.Drivers)).To(Equal(1))
				Expect(drivers.Drivers[0].Name).To(Equal("fakedriver"))
			})

			Context("when mounting given a driver name, volume id, and opaque blob of configuration", func() {
				var err error
				var volumeId string
				var mountPoint volman.MountResponse

				JustBeforeEach(func() {
					client := volhttp.NewRemoteClient(fmt.Sprintf("http://0.0.0.0:%d", volmanServerPort))
					node := GinkgoParallelNode()
					volumeId = "fake-volume_" + strconv.Itoa(node)

					testLogger.Info(fmt.Sprintf("Mounting volume: %s", volumeId))
					mountPoint, err = client.Mount(testLogger, "fakedriver", volumeId, map[string]interface{}{"volume_id": volumeId})
					Expect(err).NotTo(HaveOccurred())
				})

				It("should mount a volume", func() {
					Expect(mountPoint.Path).NotTo(Equal(""))
					defer os.Remove(mountPoint.Path)

					matches, err := filepath.Glob(mountPoint.Path)
					Expect(err).NotTo(HaveOccurred())
					Expect(len(matches)).To(Equal(1))
				})

				It("should unmount a volume given same volume ID", func() {
					client := volhttp.NewRemoteClient(fmt.Sprintf("http://0.0.0.0:%d", volmanServerPort))

					err := client.Unmount(testLogger, "fakedriver", volumeId)
					Expect(err).NotTo(HaveOccurred())

					matches, err := filepath.Glob(mountPoint.Path)
					Expect(err).NotTo(HaveOccurred())
					Expect(len(matches)).To(Equal(0))
				})
			})

			It("should error, given an invalid driver name", func() {
				client := volhttp.NewRemoteClient(fmt.Sprintf("http://0.0.0.0:%d", volmanServerPort))

				_, err := client.Mount(testLogger, "InvalidDriver", "vol", nil)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Driver 'InvalidDriver' not found in list of known drivers"))
			})

			AfterEach(func() {
				os.Remove(tmpDriversPath + "/fakedriver")
			})
		})
	})
})

func get(path string) (body string, status string, err error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://0.0.0.0:%d%s", volmanServerPort, path), nil)
	response, _ := (&http.Client{}).Do(req)
	defer response.Body.Close()
	bodyBytes, err := ioutil.ReadAll(response.Body)
	return string(bodyBytes[:]), response.Status, err
}
