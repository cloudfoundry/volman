package voldiscoverers_test

import (
	"fmt"
	"github.com/onsi/ginkgo/extensions/table"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/lager/lagertest"

	"code.cloudfoundry.org/dockerdriver"
	"code.cloudfoundry.org/dockerdriver/dockerdriverfakes"
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

		fakeDriver *dockerdriverfakes.FakeMatchableDriver

		driverName string
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("driver-discovery-test")

		fakeDriverFactory = new(volmanfakes.FakeDockerDriverFactory)

		registry = vollocal.NewPluginRegistry()
		discoverer = voldiscoverers.NewDockerDriverDiscovererWithDriverFactory(logger, registry, []string{defaultPluginsDirectory}, fakeDriverFactory)

		fakeDriver = new(dockerdriverfakes.FakeMatchableDriver)
		fakeDriver.ActivateReturns(dockerdriver.ActivateResponse{
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
				err := dockerdriver.WriteDriverSpec(logger, defaultPluginsDirectory, driverName, driverSpecExtension, driverSpecContents)
				Expect(err).NotTo(HaveOccurred())

				drivers, err = discoverer.Discover(logger)
				registry.Set(drivers)
			})

			Context("when activate returns an error", func() {
				BeforeEach(func() {
					fakeDriver.ActivateReturns(dockerdriver.ActivateResponse{Err: "Error"})
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
				})

				It("should not replace the driver in the registry", func() {
					// Expect SetDrivers not to be called
					Expect(len(drivers)).To(Equal(1))
					Expect(fakeDriverFactory.DockerDriverCallCount()).To(Equal(1))
					Expect(fakeDriver.ActivateCallCount()).To(Equal(2))
				})
				Context("when the existing driver connection is broken", func() {
					BeforeEach(func() {
						fakeDriver.ActivateReturnsOnCall(1, dockerdriver.ActivateResponse{Err: "badness"})
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

		Context("with multiple driver specs", func() {
			var (
				drivers map[string]volman.Plugin
				err     error
			)

			table.DescribeTable("should discover drivers in the order: json -> spec -> sock", func(expectedNumberOfDrivers int, specTuple ...specTuple) {
				for _, value := range specTuple {
					err := dockerdriver.WriteDriverSpec(logger, defaultPluginsDirectory, value.DriverName, value.Spec, []byte(value.SpecFileContents))
					Expect(err).NotTo(HaveOccurred())
				}

				drivers, err = discoverer.Discover(logger)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(drivers)).To(Equal(expectedNumberOfDrivers))
				Expect(fakeDriverFactory.DockerDriverCallCount()).To(Equal(expectedNumberOfDrivers))

			}, table.Entry("when there are two unique drivers with different driver formats", 2,
				jsonSpec(),
				specSpec(),
			), table.Entry("when there are three unique drivers with different driver formats", 3,
				jsonSpec(),
				specSpec(),
				sockSpec(),
			), table.Entry("when there is 1 unique driver, represented by 2 different driver formats (json and spec)", 1,
				specTuple{DriverName: "driver1", Spec: "json", SpecFileContents: `{}`},
				specTuple{DriverName: "driver1", Spec: "spec", SpecFileContents: ``},
			), table.Entry("when there is 1 unique driver, represented by 2 different driver formats (spec and sock)", 1,
				specTuple{DriverName: "driver2", Spec: "spec", SpecFileContents: ``},
				specTuple{DriverName: "driver2", Spec: "sock", SpecFileContents: ``},
			),
			)
		})

		Context("when given a compound driverspath", func() {
			BeforeEach(func() {
				discoverer = voldiscoverers.NewDockerDriverDiscovererWithDriverFactory(logger, registry, []string{defaultPluginsDirectory, secondPluginsDirectory}, fakeDriverFactory)
			})

			Context("with a single driver", func() {
				BeforeEach(func() {
					err := dockerdriver.WriteDriverSpec(logger, secondPluginsDirectory, driverName, "spec", []byte("http://0.0.0.0:8080"))
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
					err := dockerdriver.WriteDriverSpec(logger, defaultPluginsDirectory, driverName, "json", []byte("{\"Addr\":\"http://0.0.0.0:8080\"}"))
					Expect(err).NotTo(HaveOccurred())
					err = dockerdriver.WriteDriverSpec(logger, secondPluginsDirectory, "some-other-driver-name", "json", []byte("{\"Addr\":\"http://0.0.0.0:9090\"}"))
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
					err := dockerdriver.WriteDriverSpec(logger, defaultPluginsDirectory, driverName, "json", []byte("{\"Addr\":\"http://0.0.0.0:8080\"}"))
					Expect(err).NotTo(HaveOccurred())
					err = dockerdriver.WriteDriverSpec(logger, secondPluginsDirectory, driverName, "spec", []byte("http://0.0.0.0:9090"))
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
				fakeRemoteClientFactory *dockerdriverfakes.FakeRemoteClientFactory
				driverFactory           voldiscoverers.DockerDriverFactory
				fakeDriver              *dockerdriverfakes.FakeDriver
				driverDiscoverer        volman.Discoverer
			)

			JustBeforeEach(func() {
				fakeRemoteClientFactory = new(dockerdriverfakes.FakeRemoteClientFactory)
				driverFactory = voldiscoverers.NewDockerDriverFactoryWithRemoteClientFactory(fakeRemoteClientFactory)
				driverDiscoverer = voldiscoverers.NewDockerDriverDiscovererWithDriverFactory(logger, nil, []string{defaultPluginsDirectory}, driverFactory)
			})

			TestCanonicalization := func(context, actual, it, expected string) {
				Context(context, func() {
					BeforeEach(func() {
						err := dockerdriver.WriteDriverSpec(logger, defaultPluginsDirectory, driverName, "spec", []byte(actual))
						Expect(err).NotTo(HaveOccurred())
					})

					JustBeforeEach(func() {
						fakeDriver = new(dockerdriverfakes.FakeDriver)
						fakeDriver.ActivateReturns(dockerdriver.ActivateResponse{
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
					err := dockerdriver.WriteDriverSpec(logger, defaultPluginsDirectory, driverName+"2", "spec", []byte("127.0.0.1:8080"))
					err = dockerdriver.WriteDriverSpec(logger, defaultPluginsDirectory, driverName, "spec", []byte("htt%p:\\\\"))
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
				err := dockerdriver.WriteDriverSpec(logger, defaultPluginsDirectory, driverName, "spec", []byte("127.0.0.1:8080"))
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return no drivers if the driver doesn't implement VolumeDriver", func() {
				fakeDriver.ActivateReturns(dockerdriver.ActivateResponse{
					Implements: []string{"something-else"},
				})

				drivers, err := discoverer.Discover(logger)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(drivers)).To(Equal(0))
			})

			It("should return no drivers if the driver doesn't respond", func() {
				fakeDriver.ActivateReturns(dockerdriver.ActivateResponse{
					Err: "some-error",
				})

				drivers, err := discoverer.Discover(logger)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(drivers)).To(Equal(0))
			})
		})
	})
})

func sockSpec() specTuple {
	return specTuple{DriverName: "driver3", Spec: "sock", SpecFileContents: ``}
}

func specSpec() specTuple {
	return specTuple{DriverName: "driver2", Spec: "spec", SpecFileContents: `http://0.0.0.0:8080`}
}

func jsonSpec() specTuple {
	return specTuple{DriverName: "driver1", Spec: "json", SpecFileContents: `{}`}
}

type specTuple struct {
	Spec                    string
	SpecFileContents        string
	DriverName              string
	ExpectedNumberOfDrivers int
}
