package volman_test

import (
	. "github.com/cloudfoundry-incubator/volman/delegate"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ListDrivers", func() {
	BeforeEach(func() {
		client = &LocalClient{}
	})

	Context("volman has no drivers", func() {
		It("should report empty list of drivers", func() {
			drivers, err := client.ListDrivers()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(drivers.Drivers)).To(Equal(0))
		})
	})
})
