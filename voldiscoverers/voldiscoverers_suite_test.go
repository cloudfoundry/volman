package voldiscoverers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestVoldiscoverers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Voldiscoverers Suite")
}
