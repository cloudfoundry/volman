package vollocal_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-golang/clock/fakeclock"
	"github.com/pivotal-golang/lager/lagertest"

	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/cloudfoundry-incubator/volman/vollocal"
	volmanfakes "github.com/cloudfoundry-incubator/volman/volmanfakes"
	"github.com/cloudfoundry/gorouter/Godeps/_workspace/src/github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Driver Syncer", func() {
	var (
		logger *lagertest.TestLogger

		scanInterval time.Duration

		fakeClock         *fakeclock.FakeClock
		fakeDriverFactory *volmanfakes.FakeDriverFactory

		syncer  *vollocal.DriverSyncer
		process ifrit.Process
	)

	BeforeEach(func() {

		logger = lagertest.NewTestLogger("driver-syncer-test")

		fakeClock = fakeclock.NewFakeClock(time.Unix(123, 456))
		fakeDriverFactory = new(volmanfakes.FakeDriverFactory)

		scanInterval = 10 * time.Second

		syncer = vollocal.NewDriverSyncer(logger, fakeDriverFactory, scanInterval, fakeClock)
	})

	Describe("Run", func() {

		Context("when there are no drivers", func() {

			It("should have no drivers in registry map", func() {
				drivers := syncer.Drivers()
				Expect(len(drivers)).To(Equal(0))
				Expect(fakeDriverFactory.DiscoverCallCount()).To(Equal(0))
				Expect(fakeDriverFactory.DriverCallCount()).To(Equal(0))
			})

		})

		Context("when there are drivers", func() {

			BeforeEach(func() {
				fakeDriver := new(volmanfakes.FakeDriver)
				fakeDriverFactory.DiscoverReturns(map[string]voldriver.Driver{"fakedriver": fakeDriver}, nil)

				syncer = vollocal.NewDriverSyncer(logger, fakeDriverFactory, scanInterval, fakeClock)

				process = ginkgomon.Invoke(syncer)
			})

			AfterEach(func() {
				ginkgomon.Kill(process)
			})

			It("should have fake driver in registry map", func() {
				drivers := syncer.Drivers()
				Expect(len(drivers)).To(Equal(1))
				Expect(fakeDriverFactory.DiscoverCallCount()).To(Equal(1))
			})

			Context("when drivers are added", func() {

				BeforeEach(func() {
					fakeDriver := new(volmanfakes.FakeDriver)
					fakeDriverFactory.DiscoverReturns(map[string]voldriver.Driver{"anotherfakedriver": fakeDriver, "fakedriver": fakeDriver}, nil)

				})

				It("should find them!", func() {
					fakeClock.Increment(scanInterval * 2)

					Eventually(syncer.Drivers).Should(HaveLen(2))
					Expect(fakeDriverFactory.DiscoverCallCount()).To(Equal(2))
				})
			})
		})
	})
})
