package certification_test

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry-incubator/volman/certification"
	"github.com/cloudfoundry-incubator/volman/voldriver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("certification/fixture.go", func() {
	var (
		err                  error
		tmpDir, tmpFileName  string
		certificationFixture certification.CertificationFixture
	)

	BeforeEach(func() {
		tmpDir, err = ioutil.TempDir("", "certification")
		Expect(err).NotTo(HaveOccurred())

		tmpFile, err := ioutil.TempFile(tmpDir, "certification-fixture.json")
		Expect(err).NotTo(HaveOccurred())

		tmpFileName = tmpFile.Name()

		certificationFixture = certification.CertificationFixture{}
	})

	AfterEach(func() {
		err = os.RemoveAll(tmpDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("#LoadCertificationFixture", func() {
		BeforeEach(func() {
			certificationFixtureContent := `{
 						"volman_driver_path": "fake-path-to-driver",
  						"driver_name": "fakedriver",
  						"reset_driver_script": "fake-path",
						"create_config": {
						    "Name": "fake-request",
						    "Opts": {"key":"value"}
 								 }
							}`

			err = ioutil.WriteFile(tmpFileName, []byte(certificationFixtureContent), 0666)
			Expect(err).NotTo(HaveOccurred())
		})

		It("loads the fake certification fixture", func() {
			certificationFixture, err = certification.LoadCertificationFixture(tmpFileName)
			Expect(err).NotTo(HaveOccurred())

			Expect(certificationFixture.VolmanDriverPath).To(Equal("fake-path-to-driver"))
			Expect(certificationFixture.CreateConfig.Name).To(Equal("fake-request"))
		})
	})

	Context("#SaveCertificationFixture", func() {
		BeforeEach(func() {
			certificationFixture = certification.CertificationFixture{
				VolmanDriverPath:  "fake-path-to-driver",
				DriverName:        "fakedriver",
				ResetDriverScript: "fake-path",
				CreateConfig: voldriver.CreateRequest{
					Name: "fake-request",
					Opts: map[string]interface{}{"key": "value"},
				},
			}
		})

		It("saves the fake certification fixture", func() {
			err = certification.SaveCertificationFixture(certificationFixture, tmpFileName)
			Expect(err).NotTo(HaveOccurred())

			bytes, err := ioutil.ReadFile(tmpFileName)
			Expect(err).ToNot(HaveOccurred())

			readFixture := certification.CertificationFixture{}
			json.Unmarshal(bytes, &readFixture)

			Expect(readFixture.VolmanDriverPath).To(Equal("fake-path-to-driver"))
			Expect(readFixture.CreateConfig.Name).To(Equal("fake-request"))
		})
	})

})
