package voldriver_test

import (
	"encoding/json"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"

	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("SpecWriter", func() {
	var (
		err error

		testLogger = lagertest.NewTestLogger("Spec Writer")

		specWriter     voldriver.SpecWriter
		path, filename string
	)

	BeforeEach(func() {
		path, err = ioutil.TempDir(os.TempDir(), "SpecWriter")
		Expect(err).NotTo(HaveOccurred())

		specWriter = voldriver.NewSpecWriter(testLogger, "fake-driver", path)
	})

	AfterEach(func() {
		err = os.RemoveAll(path)
		Expect(err).NotTo(HaveOccurred())
	})

	Context(".json", func() {
		var (
			name, address string
		)

		BeforeEach(func() {
			filename = fmt.Sprintf("%s/%s.json", path, "fake-driver")
			name = "fake-name"
			address = "fake-address"
		})

		It("writes a .json spec file", func() {
			err := specWriter.WriteJson(voldriver.DriverSpec{
				Name:    name,
				Address: address,
			})
			Expect(err).NotTo(HaveOccurred())

			_, err = os.Stat(filename)
			Expect(os.IsNotExist(err)).To(BeFalse())

			specContents, err := ioutil.ReadFile(filename)
			Expect(err).NotTo(HaveOccurred())

			specJson := voldriver.DriverSpec{}
			err = json.Unmarshal(specContents, &specJson)
			Expect(err).NotTo(HaveOccurred())

			Expect(specJson.Name).To(Equal(name))
			Expect(specJson.Address).To(Equal(address))
		})
	})

	Context(".spec", func() {
		var address string

		BeforeEach(func() {
			filename = fmt.Sprintf("%s/%s.spec", path, "fake-driver")
			address = "fake-address"
		})

		It("writes a .spec spec file", func() {
			err = specWriter.WriteSpec(address)
			Expect(err).NotTo(HaveOccurred())

			_, err = os.Stat(filename)
			Expect(os.IsNotExist(err)).To(BeFalse())

			specContents, err := ioutil.ReadFile(filename)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(specContents)).To(ContainSubstring(address))
		})
	})
})
