package certification

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/volhttp"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var CertifyWith = func(described string, volmanFixture VolmanFixture, driverFixture DriverFixture) {
	Describe("Certify Volman with: "+described, func() {
		var (
			testLogger lager.Logger

			driverProcess ifrit.Process
			volmanProcess ifrit.Process
		)

		BeforeEach(func() {
			testLogger = lagertest.NewTestLogger("MainTest")
			volmanProcess = ginkgomon.Invoke(volmanFixture.Runner)
		})

		Context("after starting", func() {
			It("should not exit", func() {
				Consistently(volmanFixture.Runner).ShouldNot(Exit())
			})
		})

		Context("after starting volman server", func() {
			It("should get a 404 for root", func() {
				_, status, err := get("/", volmanFixture.Config.ServerPort)
				Expect(err).NotTo(HaveOccurred())
				Expect(status).Should(ContainSubstring("404"))
			})

			It("should return empty list", func() {
				client := volhttp.NewRemoteClient(fmt.Sprintf("http://0.0.0.0:%d", volmanFixture.Config.ServerPort))
				drivers, err := client.ListDrivers(testLogger)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(drivers.Drivers)).To(Equal(0))
			})

			It("should have a debug server endpoint", func() {
				_, err := http.Get(fmt.Sprintf("http://%s/debug/pprof/goroutine", volmanFixture.Config.DebugServerAddress))
				Expect(err).NotTo(HaveOccurred())
			})

			Context("after starting "+described, func() {
				BeforeEach(func() {
					driverProcess = ginkgomon.Invoke(driverFixture.Runner)
				})

				It("should return list of drivers", func() {
					client := volhttp.NewRemoteClient(fmt.Sprintf("http://0.0.0.0:%d", volmanFixture.Config.ServerPort))
					drivers, err := client.ListDrivers(testLogger)
					Expect(err).NotTo(HaveOccurred())
					Expect(len(drivers.Drivers)).To(Equal(1))
					Expect(drivers.Drivers[0].Name).To(Equal(driverFixture.Config.Name))
				})

				Context("when mounting given a driver name, volume id, and opaque blob of configuration", func() {
					var (
						err        error
						mountPoint volman.MountResponse
						client     volman.Manager
					)

					JustBeforeEach(func() {
						client = volhttp.NewRemoteClient(fmt.Sprintf("http://0.0.0.0:%d", volmanFixture.Config.ServerPort))

						testLogger.Info(fmt.Sprintf("Mounting volume: %s", driverFixture.VolumeData.Name))
						mountPoint, err = client.Mount(testLogger, driverFixture.Config.Name, driverFixture.VolumeData.Name, driverFixture.VolumeData.Config)
						Expect(err).NotTo(HaveOccurred())
					})

					It("should mount a volume", func() {
						Expect(mountPoint.Path).NotTo(Equal(""))
						defer os.Remove(mountPoint.Path)

						matches, err := filepath.Glob(mountPoint.Path)
						Expect(err).NotTo(HaveOccurred())
						Expect(len(matches)).To(Equal(1))
					})

					It("should be possible to write to the mountPoint", func() {
						testFile := path.Join(mountPoint.Path, "test.txt")
						err := ioutil.WriteFile(testFile, []byte("hello persi"), 0644)
						Expect(err).NotTo(HaveOccurred())

						err = os.Remove(testFile)
						Expect(err).NotTo(HaveOccurred())

						matches, err := filepath.Glob(mountPoint.Path + "/*")
						Expect(err).NotTo(HaveOccurred())
						Expect(len(matches)).To(Equal(0))
					})

					It("should unmount a volume given same volume ID", func() {
						err := client.Unmount(testLogger, driverFixture.Config.Name, driverFixture.VolumeData.Name)
						Expect(err).NotTo(HaveOccurred())

						matches, err := filepath.Glob(mountPoint.Path)
						Expect(err).NotTo(HaveOccurred())
						Expect(len(matches)).To(Equal(0))
					})
				})

				It("should error, given an invalid driver name", func() {
					client := volhttp.NewRemoteClient(fmt.Sprintf("http://0.0.0.0:%d", volmanFixture.Config.ServerPort))

					_, err := client.Mount(testLogger, "InvalidDriver", "vol", nil)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Driver 'InvalidDriver' not found in list of known drivers"))
				})

				AfterEach(func() {
					os.Remove(volmanFixture.Config.DriversPath + "/" + driverFixture.Config.Name)

					ginkgomon.Kill(driverProcess)

				})
			})
		})

		AfterEach(func() {
			ginkgomon.Kill(volmanProcess)
		})
	})
}

func get(path string, volmanServerPort int) (body string, status string, err error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://0.0.0.0:%d%s", volmanServerPort, path), nil)
	response, _ := (&http.Client{}).Do(req)
	defer response.Body.Close()
	bodyBytes, err := ioutil.ReadAll(response.Body)
	return string(bodyBytes[:]), response.Status, err
}
