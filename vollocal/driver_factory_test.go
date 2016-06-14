package vollocal_test

import (
	"fmt"
	"os"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/cloudfoundry-incubator/volman/vollocal"
	"github.com/cloudfoundry-incubator/volman/volmanfakes"
	"github.com/cloudfoundry/gunk/os_wrap/osfakes"
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
			driverFactory := vollocal.NewDriverFactoryWithRemoteClientFactory([]string{"some-invalid-drivers-path"}, fakeRemoteClientFactory)
			drivers, err := driverFactory.Discover(testLogger)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(drivers)).To(Equal(0))
		})
	})

	Context("when given a simple driverspath", func() {
		var (
			fakeRemoteClientFactory *volmanfakes.FakeRemoteClientFactory
			driverFactory           vollocal.DriverFactory
		)

		JustBeforeEach(func() {
			fakeRemoteClientFactory = new(volmanfakes.FakeRemoteClientFactory)
			driverFactory = vollocal.NewDriverFactoryWithRemoteClientFactory([]string{defaultPluginsDirectory}, fakeRemoteClientFactory)
		})

		Context("with a single driver", func() {

			BeforeEach(func() {
				driverName = "some-driver-name"
				err := voldriver.WriteDriverSpec(testLogger, defaultPluginsDirectory, driverName, "spec", []byte("http://0.0.0.0:8080"))
				Expect(err).NotTo(HaveOccurred())
			})

			It("should find drivers", func() {
				drivers, err := driverFactory.Discover(testLogger)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(drivers)).To(Equal(1))
				Expect(fakeRemoteClientFactory.NewRemoteClientCallCount()).To(Equal(1))
			})
		})

		Context("with hetergeneous driver specifications", func() {

			BeforeEach(func() {
				driverName = "some-driver-name"
				err := voldriver.WriteDriverSpec(testLogger, defaultPluginsDirectory, driverName, "json", []byte("{\"Addr\":\"http://0.0.0.0:8080\"}"))
				Expect(err).NotTo(HaveOccurred())
				err = voldriver.WriteDriverSpec(testLogger, defaultPluginsDirectory, driverName, "spec", []byte("http://0.0.0.0:9090"))
				Expect(err).NotTo(HaveOccurred())
			})

			It("should preferentially select spec over json specification", func() {
				_, err := driverFactory.Discover(testLogger)
				Expect(err).ToNot(HaveOccurred())
				actualAddress, _ := fakeRemoteClientFactory.NewRemoteClientArgsForCall(0)
				Expect(actualAddress).To(Equal("http://0.0.0.0:9090"))
			})
		})
	})

	Context("when a valid driver spec is discovered", func() {
		var (
			fakeRemoteClientFactory *volmanfakes.FakeRemoteClientFactory
			localDriver              *volmanfakes.FakeDriver
			driver                  voldriver.Driver
			driverFactory           vollocal.DriverFactory
		)
		BeforeEach(func() {
			driverName = "some-driver-name"
			fakeRemoteClientFactory = new(volmanfakes.FakeRemoteClientFactory)
			localDriver = new(volmanfakes.FakeDriver)
			fakeRemoteClientFactory.NewRemoteClientReturns(localDriver, nil)
			driverFactory = vollocal.NewDriverFactoryWithRemoteClientFactory([]string{defaultPluginsDirectory}, fakeRemoteClientFactory)

		})

		Context("when a json driver spec is discovered", func() {
			BeforeEach(func() {
				err := voldriver.WriteDriverSpec(testLogger, defaultPluginsDirectory, driverName, "json", []byte("{\"Addr\":\"http://0.0.0.0:8080\"}"))
				Expect(err).NotTo(HaveOccurred())
				driver, err = driverFactory.Driver(testLogger, driverName, defaultPluginsDirectory, driverName+".json")
				Expect(err).ToNot(HaveOccurred())
			})
			It("should return the correct driver", func() {
				Expect(driver).To(Equal(localDriver))
				Expect(fakeRemoteClientFactory.NewRemoteClientArgsForCall(0)).To(Equal("http://0.0.0.0:8080"))
			})
			It("should fail if unable to open file", func() {
				fakeOs := new(osfakes.FakeOs)
				driverFactory := vollocal.NewDriverFactoryWithOs([]string{defaultPluginsDirectory}, fakeOs)
				fakeOs.OpenReturns(nil, fmt.Errorf("error opening file"))
				_, err := driverFactory.Driver(testLogger, driverName, defaultPluginsDirectory, driverName+".json")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when an invalid json spec is discovered", func() {
			BeforeEach(func() {
				err := voldriver.WriteDriverSpec(testLogger, defaultPluginsDirectory, driverName, "json", []byte("{\"invalid\"}"))
				Expect(err).NotTo(HaveOccurred())
			})
			It("should error", func() {
				_, err := driverFactory.Driver(testLogger, driverName, defaultPluginsDirectory, driverName+".json")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when a spec driver spec is discovered", func() {
			BeforeEach(func() {
				err := voldriver.WriteDriverSpec(testLogger, defaultPluginsDirectory, driverName, "spec", []byte("http://0.0.0.0:8080"))
				Expect(err).NotTo(HaveOccurred())
				driver, err = driverFactory.Driver(testLogger, driverName, defaultPluginsDirectory, driverName+".spec")
				Expect(err).ToNot(HaveOccurred())
			})
			It("should return the correct driver", func() {
				Expect(driver).To(Equal(localDriver))
				Expect(fakeRemoteClientFactory.NewRemoteClientArgsForCall(0)).To(Equal("http://0.0.0.0:8080"))
			})
			It("should fail if unable to open file", func() {
				fakeOs := new(osfakes.FakeOs)
				driverFactory := vollocal.NewDriverFactoryWithOs([]string{defaultPluginsDirectory}, fakeOs)
				fakeOs.OpenReturns(nil, fmt.Errorf("error opening file"))
				_, err := driverFactory.Driver(testLogger, driverName, defaultPluginsDirectory, driverName+".spec")
				Expect(err).To(HaveOccurred())
			})

			It("should error if driver id doesn't match found driver", func() {
				fakeRemoteClientFactory := new(volmanfakes.FakeRemoteClientFactory)
				driverFactory := vollocal.NewDriverFactoryWithRemoteClientFactory([]string{defaultPluginsDirectory}, fakeRemoteClientFactory)
				_, err := driverFactory.Driver(testLogger, "garbage", defaultPluginsDirectory, "garbage.garbage")
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
				driver, err := driverFactory.Driver(testLogger, driverName, defaultPluginsDirectory, driverName+".sock")
				Expect(err).ToNot(HaveOccurred())
				Expect(driver).To(Equal(localDriver))
				address := path.Join(defaultPluginsDirectory, driverName+".sock")
				Expect(fakeRemoteClientFactory.NewRemoteClientArgsForCall(0)).To(Equal(address))
			})
			It("should error for invalid sock endpoint address", func() {
				fakeRemoteClientFactory.NewRemoteClientReturns(nil, fmt.Errorf("invalid address"))
				_, err := driverFactory.Driver(testLogger, driverName, defaultPluginsDirectory, driverName+".sock")
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
			driverFactory = vollocal.NewDriverFactoryWithRemoteClientFactory([]string{defaultPluginsDirectory}, fakeRemoteClientFactory)

		})
		It("should error", func() {
			_, err := driverFactory.Driver(testLogger, driverName, defaultPluginsDirectory, driverName+".spec")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when given a compound driverspath", func() {
		var (
			fakeRemoteClientFactory *volmanfakes.FakeRemoteClientFactory
			driverFactory           vollocal.DriverFactory
		)

		JustBeforeEach(func() {
			fakeRemoteClientFactory = new(volmanfakes.FakeRemoteClientFactory)
			driverFactory = vollocal.NewDriverFactoryWithRemoteClientFactory([]string{defaultPluginsDirectory, secondPluginsDirectory}, fakeRemoteClientFactory)
		})

		Context("with a single driver", func() {

			BeforeEach(func() {
				driverName = "some-driver-name"
				err := voldriver.WriteDriverSpec(testLogger, secondPluginsDirectory, driverName, "spec", []byte("http://0.0.0.0:8080"))
				Expect(err).NotTo(HaveOccurred())
			})

			It("should find drivers", func() {
				drivers, err := driverFactory.Discover(testLogger)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(drivers)).To(Equal(1))
				Expect(fakeRemoteClientFactory.NewRemoteClientCallCount()).To(Equal(1))
			})

		})

		Context("with multiple drivers in multiple directories", func() {

			BeforeEach(func() {
				driverName = "some-driver-name"
				err := voldriver.WriteDriverSpec(testLogger, defaultPluginsDirectory, driverName, "json", []byte("{\"Addr\":\"http://0.0.0.0:8080\"}"))
				Expect(err).NotTo(HaveOccurred())
				err = voldriver.WriteDriverSpec(testLogger, secondPluginsDirectory, "some-other-driver-name", "json", []byte("{\"Addr\":\"http://0.0.0.0:9090\"}"))
				Expect(err).NotTo(HaveOccurred())
			})

			It("should find both drivers", func() {
				drivers, err := driverFactory.Discover(testLogger)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(drivers)).To(Equal(2))
			})

		})

		Context("with the same driver but in multiple directories", func() {

			BeforeEach(func() {
				driverName = "some-driver-name"
				err := voldriver.WriteDriverSpec(testLogger, defaultPluginsDirectory, driverName, "json", []byte("{\"Addr\":\"http://0.0.0.0:8080\"}"))
				Expect(err).NotTo(HaveOccurred())
				err = voldriver.WriteDriverSpec(testLogger, secondPluginsDirectory, driverName, "spec", []byte("http://0.0.0.0:9090"))
				Expect(err).NotTo(HaveOccurred())
			})

			It("should preferentially select the driver in the first directory", func() {
				_, err := driverFactory.Discover(testLogger)
				Expect(err).ToNot(HaveOccurred())
				actualAddress, _ := fakeRemoteClientFactory.NewRemoteClientArgsForCall(0)
				Expect(actualAddress).To(Equal("http://0.0.0.0:8080"))
			})

		})

	})

	Context("when given a driver spec not in canonical form", func() {
		var (
			fakeRemoteClientFactory *volmanfakes.FakeRemoteClientFactory
			driverFactory           vollocal.DriverFactory
		)

		JustBeforeEach(func() {
			fakeRemoteClientFactory = new(volmanfakes.FakeRemoteClientFactory)
			driverFactory = vollocal.NewDriverFactoryWithRemoteClientFactory([]string{defaultPluginsDirectory}, fakeRemoteClientFactory)
		})

		TestCanonicalization := func(context, actual, it, expected string) {
			Context(context, func() {
				BeforeEach(func() {
					driverName = "some-driver-name"
					err := voldriver.WriteDriverSpec(testLogger, defaultPluginsDirectory, driverName, "spec", []byte(actual))
					Expect(err).NotTo(HaveOccurred())
				})

				It(it, func() {
					drivers, err := driverFactory.Discover(testLogger)
					Expect(err).ToNot(HaveOccurred())
					Expect(len(drivers)).To(Equal(1))
					Expect(fakeRemoteClientFactory.NewRemoteClientCallCount()).To(Equal(1))
					Expect(fakeRemoteClientFactory.NewRemoteClientArgsForCall(0)).To(Equal(expected))
				})
			})
		}

		TestCanonicalization("with an ip (and no port)", "127.0.0.1", "should return a canonicalized address", "http://127.0.0.1")
		TestCanonicalization("with an ip and port", "127.0.0.1:8080", "should return a canonicalized address", "http://127.0.0.1:8080")
		TestCanonicalization("with a tcp protocol uri with port", "tcp://127.0.0.1:8080", "should return a canonicalized address", "http://127.0.0.1:8080")
		TestCanonicalization("with a tcp protocol uri without port", "tcp://127.0.0.1", "should return a canonicalized address", "http://127.0.0.1")
		TestCanonicalization("with a unix address including protocol", "unix:///other.sock", "should return a canonicalized address", "unix:///other.sock")
		TestCanonicalization("with a unix address missing its protocol", "/other.sock", "should return a canonicalized address", "/other.sock")

		Context("with an invalid url", func() {
			BeforeEach(func() {
				driverName = "some-driver-name"
				err := voldriver.WriteDriverSpec(testLogger, defaultPluginsDirectory, driverName, "spec", []byte("htt%p:\\\\"))
				Expect(err).NotTo(HaveOccurred())
			})

			It("doesn't make a driver", func() {
				_, err := driverFactory.Discover(testLogger)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeRemoteClientFactory.NewRemoteClientCallCount()).To(Equal(0))
			})
		})
	})
})
