package fakedriver_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFakedriver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fake Driver Suite")
}
