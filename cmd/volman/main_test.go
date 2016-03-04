package main_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/volhttp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Volman", func() {

	BeforeEach(func() {
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
			testLogger := lagertest.NewTestLogger("MainTest")
			drivers, err := client.ListDrivers(testLogger)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(drivers.Drivers)).To(Equal(0))
		})
		It("should have a debug server endpoint", func() {
			_, err := http.Get(fmt.Sprintf("http://%s/debug/pprof/goroutine", debugServerAddress))
			Expect(err).NotTo(HaveOccurred())
		})
		Context("when driver installed in temp directory", func() {
			BeforeEach(func() {
				copyFile(fakeDriverPath, tmpDriversPath+"/fakedriver")
			})

			It("should return list of drivers for '/v1/drivers' (200 status)", func() {
				client := volhttp.NewRemoteClient(fmt.Sprintf("http://0.0.0.0:%d", volmanServerPort))
				testLogger := lagertest.NewTestLogger("MainTest")
				drivers, err := client.ListDrivers(testLogger)

				Expect(err).NotTo(HaveOccurred())
				Expect(len(drivers.Drivers)).To(Equal(1))
				Expect(drivers.Drivers[0].Name).To(Equal("fakedriver"))
			})

			Context("when mounting given a driver name, volume id, and opaque blob of configuration", func() {
				var err error
				var volumeId = "fake-volume"
				var mountPoint volman.MountResponse

				BeforeEach(func() {
					client := volhttp.NewRemoteClient(fmt.Sprintf("http://0.0.0.0:%d", volmanServerPort))
					testLogger := lagertest.NewTestLogger("MainTest")
					config := "Here is some config!"
					mountPoint, err = client.Mount(testLogger, "fakedriver", volumeId, config)
				})

				It("should mount a volume", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(mountPoint.Path).NotTo(Equal(""))
					defer os.Remove(mountPoint.Path)

					matches, err := filepath.Glob(mountPoint.Path)
					Expect(err).NotTo(HaveOccurred())
					Expect(len(matches)).To(Equal(1))
				})

				FIt("should unmount a volume given same volume ID", func() {
					defer os.Remove(mountPoint.Path)
					client := volhttp.NewRemoteClient(fmt.Sprintf("http://0.0.0.0:%d", volmanServerPort))
					testLogger := lagertest.NewTestLogger("MainTest")

					err := client.Unmount(testLogger, "fakedriver", volumeId)
					Expect(err).NotTo(HaveOccurred())

					// todo: actually remove path in fake driver, then enable
					// matches, err := filepath.Glob(mountPoint.Path)
					// Expect(err).NotTo(HaveOccurred())
					// Expect(len(matches)).To(Equal(0))
				})
			})

			It("should error, given an invalid driver name", func() {
				client := volhttp.NewRemoteClient(fmt.Sprintf("http://0.0.0.0:%d", volmanServerPort))
				testLogger := lagertest.NewTestLogger("MainTest")

				_, err := client.Mount(testLogger, "InvalidDriver", "vol", "")

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

func copyFile(src, dst string) error {
	err := copyFileContents(src, dst)
	if err != nil {
		return err
	}
	return copyPermissions(src, dst)
}

func copyPermissions(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, info.Mode())
}

func copyFileContents(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}

	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	return err
}
