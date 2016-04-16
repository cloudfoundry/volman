package acceptance_test

import (
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo"

	"github.com/cloudfoundry-incubator/volman/certification"
)

var _ = Describe("#CertifyWith", func() {
	fixturesPath, err := GetOrCreateFixturesPath()
	handleError(err)

	fixturesFileNames, err := filepath.Glob(fmt.Sprintf("%s/*.json", fixturesPath))
	handleError(err)

	for _, fileName := range fixturesFileNames {
		certificationFixture, err := certification.LoadCertificationFixture(fileName)
		handleError(err)

		certification.CreateRunners(&certificationFixture)
		certification.CertifyWith(certificationFixture.DriverFixture.Transport, certificationFixture.VolmanFixture, certificationFixture.DriverFixture)
	}

	It("Setting up certification fixtures", func() {})
})

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
