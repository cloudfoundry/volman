package voldocker_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestVoldocker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Voldocker Suite")
}
