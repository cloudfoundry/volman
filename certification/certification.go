package certification

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/volhttp"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var CertifyWith = func(described string, certificationFixture CertificationFixture) {
	Describe("Certify Volman with: "+described, func() {
		var (
			testLogger lager.Logger
			serverPort int = 8750

			volmanProcess ifrit.Process
			volmanRunner  ifrit.Runner
			client        volman.Manager
		)

		BeforeEach(func() {
			testLogger = lagertest.NewTestLogger("MainTest")

			if config.GinkgoConfig.ParallelTotal > 1 {
				println("DRIVER CERTIFICATION TESTS DO NOT RUN IN PARALLEL!!!")
			}
			Expect(config.GinkgoConfig.ParallelTotal).To(Equal(1))

			cmd := exec.Command("/bin/bash", certificationFixture.ResetDriverScript)
			err := cmd.Run()
			Expect(err).NotTo(HaveOccurred())

			volmanProcess = ginkgomon.Invoke(volmanRunner)

			client = volhttp.NewRemoteClient(fmt.Sprintf("http://0.0.0.0:%d", serverPort))
		})

		AfterEach(func() {
			ginkgomon.Kill(volmanProcess)
		})

		Context("after starting", func() {
			It("should not exit", func() {
				Consistently(volmanRunner).ShouldNot(Exit())
			})
		})

		Context("after starting volman server"+described, func() {
			It("should get a 404 for root", func() {
				_, status, err := get("/", serverPort)
				Expect(err).NotTo(HaveOccurred())
				Expect(status).Should(ContainSubstring("404"))
			})

			It("should return list of drivers", func() {

				drivers, err := client.ListDrivers(testLogger)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(drivers.Drivers)).ToNot(Equal(0))
			})

			It("should mount a volume", func() {

				mountPoint, err := client.Mount(testLogger, certificationFixture.DriverName, certificationFixture.CreateConfig.Name, certificationFixture.CreateConfig.Opts)
				Expect(err).NotTo(HaveOccurred())
				Expect(mountPoint.Path).NotTo(Equal(""))


				err = client.Unmount(testLogger, certificationFixture.DriverName, certificationFixture.CreateConfig.Name)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should be possible to write to the mountPoint", func() {
				mountPoint, err := client.Mount(testLogger, certificationFixture.DriverName, certificationFixture.CreateConfig.Name, certificationFixture.CreateConfig.Opts)
				Expect(err).NotTo(HaveOccurred())
				defer client.Unmount(testLogger, certificationFixture.DriverName, certificationFixture.CreateConfig.Name)

				testFile := path.Join(mountPoint.Path, "test.txt")
				err = ioutil.WriteFile(testFile, []byte("hello persi"), 0644)
				Expect(err).NotTo(HaveOccurred())

				err = os.Remove(testFile)
				Expect(err).NotTo(HaveOccurred())

				matches, err := filepath.Glob(mountPoint.Path + "/*")
				Expect(err).NotTo(HaveOccurred())
				Expect(len(matches)).To(Equal(0))

				err = client.Unmount(testLogger, certificationFixture.DriverName, certificationFixture.CreateConfig.Name)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should unmount a volume given same volume ID", func() {
				mountPoint, err := client.Mount(testLogger, certificationFixture.DriverName, certificationFixture.CreateConfig.Name, certificationFixture.CreateConfig.Opts)
				Expect(err).NotTo(HaveOccurred())

				err = client.Unmount(testLogger, certificationFixture.DriverName, certificationFixture.CreateConfig.Name)
				Expect(err).NotTo(HaveOccurred())

				matches, err := filepath.Glob(mountPoint.Path)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(matches)).To(Equal(0))
			})

			It("should error, given an invalid driver name", func() {
				_, err := client.Mount(testLogger, "InvalidDriver", "vol", nil)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Driver 'InvalidDriver' not found in list of known drivers"))
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
