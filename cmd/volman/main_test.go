package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Volman", func() {
	Context("when starting", func() {
		It("starts", func() {
			volmanProcess = ginkgomon.Invoke(runner)
			Consistently(runner).ShouldNot(Exit())
		})
	})

})
