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

			volmanFixture.CreateRunner()
			volmanProcess = ginkgomon.Invoke(volmanFixture.Runner)
		})

		AfterEach(func() {
			ginkgomon.Kill(volmanProcess)
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
					err := driverFixture.UpdateVolumeData()
					Expect(err).NotTo(HaveOccurred())

					driverFixture.CreateRunner()
					driverProcess = ginkgomon.Invoke(driverFixture.Runner)
				})

				AfterEach(func() {
					os.Remove(volmanFixture.Config.DriversPath + "/" + driverFixture.Config.Name + ".json")
					os.Remove(volmanFixture.Config.DriversPath + "/" + driverFixture.Config.Name + ".spec")
					os.Remove(volmanFixture.Config.DriversPath + "/" + driverFixture.Config.Name + ".sock")

					ginkgomon.Kill(driverProcess)
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
						client volman.Manager
					)

					JustBeforeEach(func() {
						driverFixture.UpdateVolumeData()

						client = volhttp.NewRemoteClient(fmt.Sprintf("http://0.0.0.0:%d", volmanFixture.Config.ServerPort))

						testLogger.Info(fmt.Sprintf("Mounting volume: %s", driverFixture.VolumeData.Name))

					})

					It("should mount a volume", func() {
						mountPoint, err := client.Mount(testLogger, driverFixture.Config.Name, driverFixture.VolumeData.Name, driverFixture.VolumeData.Config)
						Expect(err).NotTo(HaveOccurred())
						Expect(mountPoint.Path).NotTo(Equal(""))

						defer client.Unmount(testLogger, driverFixture.Config.Name, driverFixture.VolumeData.Name)

						matches, err := filepath.Glob(mountPoint.Path)
						Expect(err).NotTo(HaveOccurred())
						Expect(len(matches)).To(Equal(1))
					})

					It("should be possible to write to the mountPoint", func() {
						mountPoint, err := client.Mount(testLogger, driverFixture.Config.Name, driverFixture.VolumeData.Name, driverFixture.VolumeData.Config)
						Expect(err).NotTo(HaveOccurred())
						defer client.Unmount(testLogger, driverFixture.Config.Name, driverFixture.VolumeData.Name)
						testFile := path.Join(mountPoint.Path, "test.txt")
						err = ioutil.WriteFile(testFile, []byte("hello persi"), 0644)
						Expect(err).NotTo(HaveOccurred())

						err = os.Remove(testFile)
						Expect(err).NotTo(HaveOccurred())

						matches, err := filepath.Glob(mountPoint.Path + "/*")
						Expect(err).NotTo(HaveOccurred())
						Expect(len(matches)).To(Equal(0))
					})

					It("should unmount a volume given same volume ID", func() {
						mountPoint, err := client.Mount(testLogger, driverFixture.Config.Name, driverFixture.VolumeData.Name, driverFixture.VolumeData.Config)
						Expect(err).NotTo(HaveOccurred())

						err = client.Unmount(testLogger, driverFixture.Config.Name, driverFixture.VolumeData.Name)
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
			})
		})

	})
}

func get(path string, volmanServerPort int) (body string, status string, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://127.0.0.1:%d%s", volmanServerPort, path), nil)

	response, err := (&http.Client{}).Do(req)
	if err != nil {
		return "", "", err
	}

	defer response.Body.Close()
	bodyBytes, err := ioutil.ReadAll(response.Body)
	return string(bodyBytes[:]), response.Status, err
}
