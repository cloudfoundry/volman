package main_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/cloudfoundry-incubator/volman"
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
			Expect(err).ShouldNot(HaveOccurred())
			Expect(status).Should(ContainSubstring("404"))
		})
		It("should return empty list for '/v1/drivers' (200 status)", func() {
			client := volman.NewRemoteClient(fmt.Sprintf("http://0.0.0.0:%d", volmanServerPort))
			testLogger := lagertest.NewTestLogger("MainTest")
			drivers, err := client.ListDrivers(testLogger)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(drivers.Drivers)).To(Equal(0))
		})
		It("should have a debug server endpoint", func() {
			_, err := http.Get(fmt.Sprintf("http://%s/debug/pprof/goroutine", debugServerAddress))
			Expect(err).NotTo(HaveOccurred())
		})
		Context("driver installed in temp directory", func() {
			BeforeEach(func() {
				copyFile(fakeDriverPath, tmpDriversPath+"/fake_driver")
			})

			It("should return list of drivers for '/v1/drivers' (200 status)", func() {
				client := volman.NewRemoteClient(fmt.Sprintf("http://0.0.0.0:%d", volmanServerPort))
				testLogger := lagertest.NewTestLogger("MainTest")
				drivers, err := client.ListDrivers(testLogger)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(drivers.Drivers)).To(Equal(1))
				Expect(drivers.Drivers[0].Name).To(Equal("FakeDriver"))
			})

			AfterEach(func() {
				os.Remove(tmpDriversPath + "/fake_driver")
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
