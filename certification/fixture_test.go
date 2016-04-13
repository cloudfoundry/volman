package certification_test

import (
	"encoding/json"
	"github.com/cloudfoundry-incubator/volman/certification"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
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
	})

	AfterEach(func() {
		err = os.RemoveAll(tmpDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("#LoadCertificationFixture", func() {
		BeforeEach(func() {
			certificationFixtureContent := `{
				"PathToVolman": "fake-path-to-volman",
				"VolmanFixture": {
					"VolmanConfig":{
						"ServerPort": 8888,
						"DebugServerAddress": "fake-address",
						"DriversPath": "fake-drivers-path"
					}
				},
				"PathToDriver": "fake-path-to-driver",
				"MountDir": "fake-mount-dir",
				"Transport": "fake-transport",
				"ListenAddress": "fake-listen-address",
				"DriverFixture": {
					"DriverConfig": {
						"Name": "fake-name",
						"Path": "fake-path",
						"ServerPort": 5565
					},
					"VolumeData":{
						"Name": "fake-volume-name",
						"Config": {
							"fake-key0": 0,
							"fake-key1": "fake-value1"
						}
					}
				}
			}`

			err = ioutil.WriteFile(tmpFileName, []byte(certificationFixtureContent), 0666)
			Expect(err).NotTo(HaveOccurred())
		})

		It("loads the fake certification fixture", func() {
			certificationFixture, err = certification.LoadCertificationFixture(tmpFileName)
			Expect(err).NotTo(HaveOccurred())

			Expect(certificationFixture.PathToVolman).To(Equal("fake-path-to-volman"))
			Expect(certificationFixture.VolmanFixture.Config.ServerPort).To(Equal(8888))

			Expect(certificationFixture.PathToDriver).To(Equal("fake-path-to-driver"))
			Expect(certificationFixture.MountDir).To(Equal("fake-mount-dir"))
			Expect(certificationFixture.Transport).To(Equal("fake-transport"))
			Expect(certificationFixture.ListenAddress).To(Equal("fake-listen-address"))
			Expect(certificationFixture.DriverFixture.Config.Name).To(Equal("fake-name"))
			Expect(certificationFixture.DriverFixture.VolumeData.Name).To(Equal("fake-volume-name"))
		})
	})

	Context("#SaveCertificationFixture", func() {
		BeforeEach(func() {
			certificationFixture = certification.CertificationFixture{
				PathToVolman: "fake-path-to-volman",
				VolmanFixture: certification.VolmanFixture{
					Config: certification.VolmanConfig{
						ServerPort:         8888,
						DebugServerAddress: "fake-address",
						DriversPath:        "fake-drivers-path",
					},
				},
				PathToDriver:  "fake-path-to-driver",
				MountDir:      "fake-mount-dir",
				Transport:     "fake-transport",
				ListenAddress: "fake-listen-address",
				DriverFixture: certification.DriverFixture{
					Config: certification.DriverConfig{
						Name:       "fake-name",
						Path:       "fake-path",
						ServerPort: 5565,
					},
					VolumeData: certification.VolumeData{
						Name: "fake-volume-name",
						Config: map[string]interface{}{
							"fake-key0": 0,
							"fake-key1": "fake-value1",
						},
					},
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

			Expect(readFixture.PathToVolman).To(Equal("fake-path-to-volman"))
			Expect(readFixture.VolmanFixture.Config.ServerPort).To(Equal(8888))

			Expect(readFixture.PathToDriver).To(Equal("fake-path-to-driver"))
			Expect(readFixture.MountDir).To(Equal("fake-mount-dir"))
			Expect(readFixture.Transport).To(Equal("fake-transport"))
			Expect(readFixture.ListenAddress).To(Equal("fake-listen-address"))
			Expect(readFixture.DriverFixture.Config.Name).To(Equal("fake-name"))
			Expect(readFixture.DriverFixture.VolumeData.Name).To(Equal("fake-volume-name"))
		})
	})
})
