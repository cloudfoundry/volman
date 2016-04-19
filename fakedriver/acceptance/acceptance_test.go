package acceptance_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"

	"github.com/cloudfoundry-incubator/volman/certification"
)

var _ = Describe("#CertifyWith", func() {
	fixturesFileNames, err := filepath.Glob("../../fixtures/*.json")
	if err != nil {
		panic(err)
	}

	for _, fileName := range fixturesFileNames {
		certificationFixture, err := certification.LoadCertificationFixture(fileName)
		if err != nil {
			panic(err)
		}
		certification.CertifyWith("fake driver", certificationFixture)
	}
	It("Setting up certification fixtures", func() {})

})
