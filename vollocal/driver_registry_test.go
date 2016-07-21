package vollocal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry-incubator/volman/vollocal"

	"github.com/cloudfoundry-incubator/voldriver"
	"github.com/cloudfoundry-incubator/voldriver/voldriverfakes"
)

var _ = Describe("DriverRegistry", func() {
	var (
		emptyRegistry, oneRegistry, manyRegistry DriverRegistry
	)

	BeforeEach(func() {
		emptyRegistry = NewDriverRegistry()

		oneRegistry = NewDriverRegistryWith(map[string]voldriver.Driver{
			"one": new(voldriverfakes.FakeDriver),
		})

		manyRegistry = NewDriverRegistryWith(map[string]voldriver.Driver{
			"one": new(voldriverfakes.FakeDriver),
			"two": new(voldriverfakes.FakeDriver),
		})
	})

	Describe("#Driver", func() {
		It("sets the driver to new value", func() {
			oneDriver, exists := oneRegistry.Driver("one")
			Expect(exists).To(BeTrue())
			Expect(oneDriver).NotTo(BeNil())
		})

		It("returns nil and false if the driver doesn't exist", func() {
			oneDriver, exists := oneRegistry.Driver("doesnotexist")
			Expect(exists).To(BeFalse())
			Expect(oneDriver).To(BeNil())
		})
	})

	Describe("#Activate", func() {
		It("returns true when driver is activated", func() {
			activated, err := oneRegistry.Activated("one")
			Expect(activated).To(BeFalse())
			Expect(err).NotTo(HaveOccurred())

			err = oneRegistry.Activate("one")
			Expect(err).NotTo(HaveOccurred())

			activated, err = oneRegistry.Activated("one")
			Expect(activated).To(BeTrue())
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns false and error for non-present drivers", func() {
			activated, err := oneRegistry.Activated("two")
			Expect(err).To(HaveOccurred())
			Expect(activated).To(BeFalse())
		})
	})

	Describe("#Drivers", func() {
		It("should return return empty map for emptyRegistry", func() {
			drivers := emptyRegistry.Drivers()
			Expect(len(drivers)).To(Equal(0))
		})

		It("should return return one driver for oneRegistry", func() {
			drivers := oneRegistry.Drivers()
			Expect(len(drivers)).To(Equal(1))
		})
	})

	Describe("#Add", func() {
		It("fails when adding driver that already exists", func() {
			newDriver := new(voldriverfakes.FakeDriver)
			err := oneRegistry.Add("one", newDriver)
			Expect(err).To(HaveOccurred())
		})

		It("adds driver that does not exists", func() {
			newDriver := new(voldriverfakes.FakeDriver)
			err := manyRegistry.Add("three", newDriver)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("#Keys", func() {
		It("should return return {'one'} for oneRegistry keys", func() {
			keys := emptyRegistry.Keys()
			Expect(len(keys)).To(Equal(0))
		})

		It("should return return {'one'} for oneRegistry keys", func() {
			keys := oneRegistry.Keys()
			Expect(len(keys)).To(Equal(1))
			Expect(keys[0]).To(Equal("one"))
		})
	})
})
