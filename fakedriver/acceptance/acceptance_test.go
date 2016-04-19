package acceptance_test

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"

	"github.com/cloudfoundry-incubator/volman/certification"
)

var _ = Describe("#CertifyWith", func() {

	fixturesPath := os.Getenv("FIXTURES_PATH")
	fixturesFileNames, err := filepath.Glob(fmt.Sprintf("%s/*.json", fixturesPath))
	handleError(err)

	for _, fileName := range fixturesFileNames {
		certificationFixture, err := certification.LoadCertificationFixture(fileName)
		handleError(err)
		certification.CertifyWith("fake driver", certificationFixture)
	}
	It("Setting up certification fixtures", func() {})

})

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

// To run tests, setup up your driver, place your specs in the driver path

/*
{
	"volman_bin_path":"/tmp/builds/volman",
	"volman_driver_path":"/tmp/plugins",
	"driver_name": "fakedriver",
	"reset_driver_script":"/tmp/resetScripts/reset.sh",
	"valid_create_request": {
		"name": "fakedriver",
		"opts": {
			"keyring":"lalala",
		}
	}
}
*/
