package voldiscoverers_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/lager/lagertest"

	"code.cloudfoundry.org/voldriver"
	"code.cloudfoundry.org/voldriver/voldriverfakes"
	"code.cloudfoundry.org/volman"
	"code.cloudfoundry.org/volman/voldiscoverers"
	"code.cloudfoundry.org/volman/vollocal"
	"code.cloudfoundry.org/volman/volmanfakes"
)

var _ = Describe("Docker Driver Discoverer", func() {
	var (
		logger *lagertest.TestLogger

		fakeDriverFactory *volmanfakes.FakeDockerDriverFactory

		registry   volman.PluginRegistry
		discoverer volman.Discoverer

		fakeDriver *voldriverfakes.FakeMatchableDriver

		driverName string
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("driver-discovery-test")

		fakeDriverFactory = new(volmanfakes.FakeDockerDriverFactory)

		registry = vollocal.NewPluginRegistry()
		discoverer = voldiscoverers.NewDockerDriverDiscovererWithDriverFactory(logger, registry, []string{defaultPluginsDirectory}, fakeDriverFactory)

		fakeDriver = new(voldriverfakes.FakeMatchableDriver)
		fakeDriver.ActivateReturns(voldriver.ActivateResponse{
			Implements: []string{"VolumeDriver"},
		})

		fakeDriverFactory.DockerDriverReturns(fakeDriver, nil)

		driverName = fmt.Sprintf("fakedriver-%d", GinkgoConfig.ParallelNode)
	})

	Describe("#Discover", func() {
		Context("when given driverspath with no drivers", func() {
			It("no drivers are found", func() {
				drivers, err := discoverer.Discover(logger)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(drivers)).To(Equal(0))
			})
		})

		Context("with a single driver", func() {
			var (
				drivers             map[string]volman.Plugin
				err                 error
				driverSpecContents  []byte
				driverSpecExtension string
			)

			BeforeEach(func() {
				driverSpecContents = []byte("http://0.0.0.0:8080")
				driverSpecExtension = "spec"
			})

			JustBeforeEach(func() {
				err := voldriver.WriteDriverSpec(logger, defaultPluginsDirectory, driverName, driverSpecExtension, driverSpecContents)
				Expect(err).NotTo(HaveOccurred())

				drivers, err = discoverer.Discover(logger)
				registry.Set(drivers)
			})

			Context("when activate returns an error", func() {
				BeforeEach(func() {
					fakeDriver.ActivateReturns(voldriver.ActivateResponse{Err: "Error"})
				})
				It("should not find drivers that are unresponsive", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(len(drivers)).To(Equal(0))
					Expect(fakeDriverFactory.DockerDriverCallCount()).To(Equal(1))
				})
			})

			It("should find drivers", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(len(drivers)).To(Equal(1))
				Expect(fakeDriverFactory.DockerDriverCallCount()).To(Equal(1))
			})

			Context("when discover is running with the same config", func() {
				BeforeEach(func() {
					fakeDriver.MatchesReturns(true)
				})

				JustBeforeEach(func() {
					Expect(len(drivers)).To(Equal(1))
					Expect(fakeDriverFactory.DockerDriverCallCount()).To(Equal(1))

					drivers, err = discoverer.Discover(logger)
					registry.Set(drivers)
				})

				It("should not replace the driver in the registry", func() {
					// Expect SetDrivers not to be called
					Expect(len(drivers)).To(Equal(1))
					Expect(fakeDriverFactory.DockerDriverCallCount()).To(Equal(1))
					Expect(fakeDriver.ActivateCallCount()).To(Equal(2))
				})
				Context("when the existing driver connection is broken", func() {
					BeforeEach(func() {
						fakeDriver.ActivateReturnsOnCall(1, voldriver.ActivateResponse{Err: "badness"})
					})
					It("should replace the driver in the registry", func() {
						Expect(len(drivers)).To(Equal(1))
						Expect(fakeDriverFactory.DockerDriverCallCount()).To(Equal(2))
						Expect(fakeDriver.ActivateCallCount()).To(Equal(3))
					})
				})
			})

			Context("with different config", func() {
				BeforeEach(func() {
					fakeDriver.MatchesReturns(false)
				})

				JustBeforeEach(func() {
					Expect(len(drivers)).To(Equal(1))
					Expect(fakeDriverFactory.DockerDriverCallCount()).To(Equal(1))

					drivers, err = discoverer.Discover(logger)
					registry.Set(drivers)
				})

				It("should replace the driver in the registry", func() {
					// Expect SetDrivers to be called
					Expect(len(drivers)).To(Equal(1))
					Expect(fakeDriverFactory.DockerDriverCallCount()).To(Equal(2))
					Expect(fakeDriver.ActivateCallCount()).To(Equal(2))
				})
			})

			Context("when the driver opts in to unique volume IDs", func() {
				BeforeEach(func() {
					driverSpecContents = []byte("{\"Addr\":\"http://0.0.0.0:8080\",\"UniqueVolumeIds\": true}")
					driverSpecExtension = "json"
				})

				It("discovers the driver and sets the flag on the plugin state", func() {
					Expect(err).ToNot(HaveOccurred())

					Expect(len(drivers)).To(Equal(1))
					plugin := drivers[driverName]
					Expect(plugin.GetPluginSpec().UniqueVolumeIds).To(BeTrue())

					Expect(fakeDriverFactory.DockerDriverCallCount()).To(Equal(1))
				})
			})
		})

		Context("when given a simple driverspath", func() {
			Context("with hetergeneous json+spec driver specifications", func() {
				BeforeEach(func() {
					err := voldriver.WriteDriverSpec(logger, defaultPluginsDirectory, driverName, "json", []byte("{\"Addr\":\"http://0.0.0.0:8080\"}"))
					Expect(err).NotTo(HaveOccurred())
					err = voldriver.WriteDriverSpec(logger, defaultPluginsDirectory, driverName, "spec", []byte("http://0.0.0.0:9090"))
					Expect(err).NotTo(HaveOccurred())
				})

				It("should preferentially select json over spec specification", func() {
					drivers, err := discoverer.Discover(logger)
					Expect(err).ToNot(HaveOccurred())
					Expect(len(drivers)).To(Equal(1))
					_, _, _, specFileName := fakeDriverFactory.DockerDriverArgsForCall(0)
					Expect(specFileName).To(Equal(driverName + ".json"))
				})
			})

			Context("with hetergeneous spec+sock driver specifications", func() {
				BeforeEach(func() {
					err := voldriver.WriteDriverSpec(logger, defaultPluginsDirectory, driverName, "sock", []byte("unix:///some.sock"))
					Expect(err).NotTo(HaveOccurred())
					err = voldriver.WriteDriverSpec(logger, defaultPluginsDirectory, driverName, "spec", []byte("http://0.0.0.0:9090"))
					Expect(err).NotTo(HaveOccurred())
				})

				It("should preferentially select spec over sock specification", func() {
					drivers, err := discoverer.Discover(logger)
					Expect(err).ToNot(HaveOccurred())
					Expect(len(drivers)).To(Equal(1))
					_, _, _, specFileName := fakeDriverFactory.DockerDriverArgsForCall(0)
					Expect(specFileName).To(Equal(driverName + ".spec"))
				})
			})
		})

		Context("when given a compound driverspath", func() {
			BeforeEach(func() {
				discoverer = voldiscoverers.NewDockerDriverDiscovererWithDriverFactory(logger, registry, []string{defaultPluginsDirectory, secondPluginsDirectory}, fakeDriverFactory)
			})

			Context("with a single driver", func() {
				BeforeEach(func() {
					err := voldriver.WriteDriverSpec(logger, secondPluginsDirectory, driverName, "spec", []byte("http://0.0.0.0:8080"))
					Expect(err).NotTo(HaveOccurred())
				})

				It("should find drivers", func() {
					drivers, err := discoverer.Discover(logger)
					Expect(err).ToNot(HaveOccurred())
					Expect(len(drivers)).To(Equal(1))
					Expect(fakeDriverFactory.DockerDriverCallCount()).To(Equal(1))
				})

			})

			Context("with multiple drivers in multiple directories", func() {
				BeforeEach(func() {
					err := voldriver.WriteDriverSpec(logger, defaultPluginsDirectory, driverName, "json", []byte("{\"Addr\":\"http://0.0.0.0:8080\"}"))
					Expect(err).NotTo(HaveOccurred())
					err = voldriver.WriteDriverSpec(logger, secondPluginsDirectory, "some-other-driver-name", "json", []byte("{\"Addr\":\"http://0.0.0.0:9090\"}"))
					Expect(err).NotTo(HaveOccurred())
				})

				It("should find both drivers", func() {
					drivers, err := discoverer.Discover(logger)
					Expect(err).ToNot(HaveOccurred())
					Expect(len(drivers)).To(Equal(2))
				})
			})

			Context("with the same driver but in multiple directories", func() {
				BeforeEach(func() {
					err := voldriver.WriteDriverSpec(logger, defaultPluginsDirectory, driverName, "json", []byte("{\"Addr\":\"http://0.0.0.0:8080\"}"))
					Expect(err).NotTo(HaveOccurred())
					err = voldriver.WriteDriverSpec(logger, secondPluginsDirectory, driverName, "spec", []byte("http://0.0.0.0:9090"))
					Expect(err).NotTo(HaveOccurred())
				})

				It("should preferentially select the driver in the first directory", func() {
					_, err := discoverer.Discover(logger)
					Expect(err).ToNot(HaveOccurred())
					_, _, _, specFileName := fakeDriverFactory.DockerDriverArgsForCall(0)
					Expect(specFileName).To(Equal(driverName + ".json"))
				})
			})
		})

		Context("when given a driver spec not in canonical form", func() {
			var (
				fakeRemoteClientFactory *voldriverfakes.FakeRemoteClientFactory
				driverFactory           voldiscoverers.DockerDriverFactory
				fakeDriver              *voldriverfakes.FakeDriver
				driverDiscoverer        volman.Discoverer
			)

			JustBeforeEach(func() {
				fakeRemoteClientFactory = new(voldriverfakes.FakeRemoteClientFactory)
				driverFactory = voldiscoverers.NewDockerDriverFactoryWithRemoteClientFactory(fakeRemoteClientFactory)
				driverDiscoverer = voldiscoverers.NewDockerDriverDiscovererWithDriverFactory(logger, nil, []string{defaultPluginsDirectory}, driverFactory)
			})

			TestCanonicalization := func(context, actual, it, expected string) {
				Context(context, func() {
					BeforeEach(func() {
						err := voldriver.WriteDriverSpec(logger, defaultPluginsDirectory, driverName, "spec", []byte(actual))
						Expect(err).NotTo(HaveOccurred())
					})

					JustBeforeEach(func() {
						fakeDriver = new(voldriverfakes.FakeDriver)
						fakeDriver.ActivateReturns(voldriver.ActivateResponse{
							Implements: []string{"VolumeDriver"},
						})

						fakeRemoteClientFactory.NewRemoteClientReturns(fakeDriver, nil)
					})

					It(it, func() {
						drivers, err := driverDiscoverer.Discover(logger)
						Expect(err).ToNot(HaveOccurred())
						Expect(len(drivers)).To(Equal(1))
						Expect(fakeRemoteClientFactory.NewRemoteClientCallCount()).To(Equal(1))
						Expect(fakeRemoteClientFactory.NewRemoteClientArgsForCall(0)).To(Equal(expected))
					})
				})
			}

			TestCanonicalization("with an ip (and no port)", "127.0.0.1", "should return a canonicalized address", "http://127.0.0.1")
			TestCanonicalization("with a tcp protocol uri with port", "tcp://127.0.0.1:8080", "should return a canonicalized address", "http://127.0.0.1:8080")
			TestCanonicalization("with a tcp protocol uri without port", "tcp://127.0.0.1", "should return a canonicalized address", "http://127.0.0.1")
			TestCanonicalization("with a unix address including protocol", "unix:///other.sock", "should return a canonicalized address", "unix:///other.sock")
			TestCanonicalization("with a unix address missing its protocol", "/other.sock", "should return a canonicalized address", "/other.sock")

			Context("with an invalid url", func() {
				BeforeEach(func() {
					err := voldriver.WriteDriverSpec(logger, defaultPluginsDirectory, driverName+"2", "spec", []byte("127.0.0.1:8080"))
					err = voldriver.WriteDriverSpec(logger, defaultPluginsDirectory, driverName, "spec", []byte("htt%p:\\\\"))
					Expect(err).NotTo(HaveOccurred())
				})

				It("doesn't make a driver", func() {
					_, err := driverDiscoverer.Discover(logger)
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeRemoteClientFactory.NewRemoteClientCallCount()).To(Equal(0))
				})
			})
		})

		Context("when given a driver spec with a bad driver", func() {
			BeforeEach(func() {
				err := voldriver.WriteDriverSpec(logger, defaultPluginsDirectory, driverName, "spec", []byte("127.0.0.1:8080"))
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return no drivers if the driver doesn't implement VolumeDriver", func() {
				fakeDriver.ActivateReturns(voldriver.ActivateResponse{
					Implements: []string{"something-else"},
				})

				drivers, err := discoverer.Discover(logger)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(drivers)).To(Equal(0))
			})

			It("should return no drivers if the driver doesn't respond", func() {
				fakeDriver.ActivateReturns(voldriver.ActivateResponse{
					Err: "some-error",
				})

				drivers, err := discoverer.Discover(logger)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(drivers)).To(Equal(0))
			})
		})
	})
})
