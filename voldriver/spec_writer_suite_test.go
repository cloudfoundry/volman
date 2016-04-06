package voldriver_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSpecWriter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Driver SpecWriter Suite")
}
