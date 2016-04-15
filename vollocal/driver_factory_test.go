package vollocal_test

import (
	"os"
	"path"

	"fmt"

	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/cloudfoundry-incubator/volman/vollocal"
	"github.com/cloudfoundry-incubator/volman/volmanfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("DriverFactory", func() {
	var (
		testLogger lager.Logger
		driverName string
	)
	BeforeEach(func() {
		testLogger = lagertest.NewTestLogger("ClientTest")
	})
	Context("when given driverspath with no drivers", func() {
		It("no drivers are found", func() {
			fakeRemoteClientFactory := new(volmanfakes.FakeRemoteClientFactory)
			driverFactory := vollocal.NewDriverFactoryWithRemoteClientFactory("some-invalid-drivers-path", fakeRemoteClientFactory)
			drivers, err := driverFactory.Discover(testLogger)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(drivers)).To(Equal(0))
		})
	})
	Context("when given driverspath with a single driver", func() {

		BeforeEach(func() {
			driverName = "some-driver-name"
			err := voldriver.WriteDriverSpec(testLogger, defaultPluginsDirectory, driverName, "http://0.0.0.0:8080")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should find drivers", func() {
			fakeRemoteClientFactory := new(volmanfakes.FakeRemoteClientFactory)
			driverFactory := vollocal.NewDriverFactoryWithRemoteClientFactory(defaultPluginsDirectory, fakeRemoteClientFactory)
			drivers, err := driverFactory.Discover(testLogger)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(drivers)).To(Equal(1))
			Expect(fakeRemoteClientFactory.NewRemoteClientCallCount()).To(Equal(1))
		})
	})
	Context("when given driverspath with hetergeneous driver specifications", func() {

		BeforeEach(func() {
			driverName = "some-driver-name"
			err := voldriver.WriteDriverJSONSpec(testLogger, defaultPluginsDirectory, driverName, []byte("{\"url\":\"http://0.0.0.0:8080\"}"))
			Expect(err).NotTo(HaveOccurred())
			err = voldriver.WriteDriverSpec(testLogger, defaultPluginsDirectory, driverName, "http://0.0.0.0:9090")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should preferentially select spec over json specification", func() {
			fakeRemoteClientFactory := new(volmanfakes.FakeRemoteClientFactory)
			driverFactory := vollocal.NewDriverFactoryWithRemoteClientFactory(defaultPluginsDirectory, fakeRemoteClientFactory)
			_, err := driverFactory.Discover(testLogger)
			Expect(err).ToNot(HaveOccurred())
			actualAddress := fakeRemoteClientFactory.NewRemoteClientArgsForCall(0)
			Expect(actualAddress).To(Equal("http://0.0.0.0:9090"))
		})

	})
	Context("when a valid driver spec is discovered", func() {
		var (
			fakeRemoteClientFactory *volmanfakes.FakeRemoteClientFactory
			fakeDriver              *volmanfakes.FakeDriver
			driver                  voldriver.Driver
			driverFactory           vollocal.DriverFactory
		)
		BeforeEach(func() {
			driverName = "some-driver-name"
			fakeRemoteClientFactory = new(volmanfakes.FakeRemoteClientFactory)
			fakeDriver = new(volmanfakes.FakeDriver)
			fakeRemoteClientFactory.NewRemoteClientReturns(fakeDriver, nil)
			driverFactory = vollocal.NewDriverFactoryWithRemoteClientFactory(defaultPluginsDirectory, fakeRemoteClientFactory)

		})

		Context("when a json driver spec is discovered", func() {
			BeforeEach(func() {
				err := voldriver.WriteDriverJSONSpec(testLogger, defaultPluginsDirectory, driverName, []byte("{\"Addr\":\"http://0.0.0.0:8080\"}"))
				Expect(err).NotTo(HaveOccurred())
				driver, err = driverFactory.Driver(testLogger, driverName, driverName+".json")
				Expect(err).ToNot(HaveOccurred())
			})
			It("should return the correct driver", func() {
				Expect(driver).To(Equal(fakeDriver))
				Expect(fakeRemoteClientFactory.NewRemoteClientArgsForCall(0)).To(Equal("http://0.0.0.0:8080"))
			})
			It("should fail if unable to open file", func() {
				fakeOs := new(volmanfakes.FakeOs)
				driverFactory := vollocal.NewDriverFactoryWithOs(defaultPluginsDirectory, fakeOs)
				fakeOs.OpenReturns(nil, fmt.Errorf("error opening file"))
				_, err := driverFactory.Driver(testLogger, driverName, driverName+".json")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when an invalid json spec is discovered", func() {
			BeforeEach(func() {
				err := voldriver.WriteDriverJSONSpec(testLogger, defaultPluginsDirectory, driverName, []byte("{\"invalid\"}"))
				Expect(err).NotTo(HaveOccurred())
			})
			It("should error", func() {
				_, err := driverFactory.Driver(testLogger, driverName, driverName+".json")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when a spec driver spec is discovered", func() {
			BeforeEach(func() {
				err := voldriver.WriteDriverSpec(testLogger, defaultPluginsDirectory, driverName, "http://0.0.0.0:8080")
				Expect(err).NotTo(HaveOccurred())
				driver, err = driverFactory.Driver(testLogger, driverName, driverName+".spec")
				Expect(err).ToNot(HaveOccurred())
			})
			It("should return the correct driver", func() {
				Expect(driver).To(Equal(fakeDriver))
				Expect(fakeRemoteClientFactory.NewRemoteClientArgsForCall(0)).To(Equal("http://0.0.0.0:8080"))
			})
			It("should fail if unable to open file", func() {
				fakeOs := new(volmanfakes.FakeOs)
				driverFactory := vollocal.NewDriverFactoryWithOs(defaultPluginsDirectory, fakeOs)
				fakeOs.OpenReturns(nil, fmt.Errorf("error opening file"))
				_, err := driverFactory.Driver(testLogger, driverName, driverName+".spec")
				Expect(err).To(HaveOccurred())
			})

			It("should error if driver id doesn't match found driver", func() {
				fakeRemoteClientFactory := new(volmanfakes.FakeRemoteClientFactory)
				driverFactory := vollocal.NewDriverFactoryWithRemoteClientFactory(defaultPluginsDirectory, fakeRemoteClientFactory)
				_, err := driverFactory.Driver(testLogger, "garbage", "garbage.garbage")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when a sock driver spec is discovered", func() {
			BeforeEach(func() {
				f, err := os.Create(defaultPluginsDirectory + "/" + driverName + ".sock")
				defer f.Close()
				Expect(err).ToNot(HaveOccurred())
			})
			It("should return the correct driver", func() {
				driver, err := driverFactory.Driver(testLogger, driverName, driverName+".sock")
				Expect(err).ToNot(HaveOccurred())
				Expect(driver).To(Equal(fakeDriver))
				address := path.Join(defaultPluginsDirectory, driverName+".sock")
				Expect(fakeRemoteClientFactory.NewRemoteClientArgsForCall(0)).To(Equal(address))
			})
			It("should error for invalid sock endpoint address", func() {
				fakeRemoteClientFactory.NewRemoteClientReturns(nil, fmt.Errorf("invalid address"))
				_, err := driverFactory.Driver(testLogger, driverName, driverName+".sock")
				Expect(err).To(HaveOccurred())
			})
		})
	})
	Context("when valid driver spec is not discovered", func() {
		var (
			fakeRemoteClientFactory *volmanfakes.FakeRemoteClientFactory
			fakeDriver              *volmanfakes.FakeDriver
			driverFactory           vollocal.DriverFactory
		)
		BeforeEach(func() {
			driverName = "some-driver-name"
			fakeRemoteClientFactory = new(volmanfakes.FakeRemoteClientFactory)
			fakeDriver = new(volmanfakes.FakeDriver)
			fakeRemoteClientFactory.NewRemoteClientReturns(fakeDriver, nil)
			driverFactory = vollocal.NewDriverFactoryWithRemoteClientFactory(defaultPluginsDirectory, fakeRemoteClientFactory)

		})
		It("should error", func() {
			_, err := driverFactory.Driver(testLogger, driverName, driverName+".spec")
			Expect(err).To(HaveOccurred())
		})
	})
})
