package voldiscoverers_test

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/goshims/osshim/os_fake"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/voldriver"
	"code.cloudfoundry.org/voldriver/voldriverfakes"
	"code.cloudfoundry.org/volman/voldiscoverers"
)

var _ = Describe("DriverFactory", func() {
	var (
		testLogger lager.Logger
		driverName string
	)
	BeforeEach(func() {
		testLogger = lagertest.NewTestLogger("ClientTest")
	})

	Context("when a valid driver spec is discovered", func() {
		var (
			fakeRemoteClientFactory *voldriverfakes.FakeRemoteClientFactory
			localDriver             *voldriverfakes.FakeDriver
			driver                  voldriver.Driver
			driverFactory           voldiscoverers.DockerDriverFactory
		)
		BeforeEach(func() {
			driverName = "some-driver-name"
			fakeRemoteClientFactory = new(voldriverfakes.FakeRemoteClientFactory)
			localDriver = new(voldriverfakes.FakeDriver)
			fakeRemoteClientFactory.NewRemoteClientReturns(localDriver, nil)
			driverFactory = voldiscoverers.NewDockerDriverFactoryWithRemoteClientFactory(fakeRemoteClientFactory)

		})

		Context("when a json driver spec is discovered", func() {
			BeforeEach(func() {
				err := voldriver.WriteDriverSpec(testLogger, defaultPluginsDirectory, driverName, "json", []byte("{\"Addr\":\"http://0.0.0.0:8080\"}"))
				Expect(err).NotTo(HaveOccurred())
				driver, err = driverFactory.DockerDriver(testLogger, driverName, defaultPluginsDirectory, driverName+".json")
				Expect(err).ToNot(HaveOccurred())
			})
			It("should return the correct driver", func() {
				Expect(driver).To(Equal(localDriver))
				Expect(fakeRemoteClientFactory.NewRemoteClientArgsForCall(0)).To(Equal("http://0.0.0.0:8080"))
			})
			It("should fail if unable to open file", func() {
				fakeOs := new(os_fake.FakeOs)
				driverFactory := voldiscoverers.NewDockerDriverFactoryWithOs(fakeOs)
				fakeOs.OpenReturns(nil, fmt.Errorf("error opening file"))
				_, err := driverFactory.DockerDriver(testLogger, driverName, defaultPluginsDirectory, driverName+".json")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when an invalid json spec is discovered", func() {
			BeforeEach(func() {
				err := voldriver.WriteDriverSpec(testLogger, defaultPluginsDirectory, driverName, "json", []byte("{\"invalid\"}"))
				Expect(err).NotTo(HaveOccurred())
			})
			It("should error", func() {
				_, err := driverFactory.DockerDriver(testLogger, driverName, defaultPluginsDirectory, driverName+".json")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when a spec driver spec is discovered", func() {
			BeforeEach(func() {
				err := voldriver.WriteDriverSpec(testLogger, defaultPluginsDirectory, driverName, "spec", []byte("http://0.0.0.0:8080"))
				Expect(err).NotTo(HaveOccurred())
				driver, err = driverFactory.DockerDriver(testLogger, driverName, defaultPluginsDirectory, driverName+".spec")
				Expect(err).ToNot(HaveOccurred())
			})
			It("should return the correct driver", func() {
				Expect(driver).To(Equal(localDriver))
				Expect(fakeRemoteClientFactory.NewRemoteClientArgsForCall(0)).To(Equal("http://0.0.0.0:8080"))
			})
			It("should fail if unable to open file", func() {
				fakeOs := new(os_fake.FakeOs)
				driverFactory := voldiscoverers.NewDockerDriverFactoryWithOs(fakeOs)
				fakeOs.OpenReturns(nil, fmt.Errorf("error opening file"))
				_, err := driverFactory.DockerDriver(testLogger, driverName, defaultPluginsDirectory, driverName+".spec")
				Expect(err).To(HaveOccurred())
			})

			It("should error if driver id doesn't match found driver", func() {
				fakeRemoteClientFactory := new(voldriverfakes.FakeRemoteClientFactory)
				driverFactory := voldiscoverers.NewDockerDriverFactoryWithRemoteClientFactory(fakeRemoteClientFactory)
				_, err := driverFactory.DockerDriver(testLogger, "garbage", defaultPluginsDirectory, "garbage.garbage")
				Expect(err).To(HaveOccurred())
			})
		})

		if runtime.GOOS != "windows" {
			Context("when a sock driver spec is discovered", func() {
				BeforeEach(func() {
					f, err := os.Create(filepath.Join(defaultPluginsDirectory, driverName+".sock"))
					defer f.Close()
					Expect(err).ToNot(HaveOccurred())
				})
				It("should return the correct driver", func() {
					driver, err := driverFactory.DockerDriver(testLogger, driverName, defaultPluginsDirectory, driverName+".sock")
					Expect(err).ToNot(HaveOccurred())
					Expect(driver).To(Equal(localDriver))
					address := path.Join(defaultPluginsDirectory, driverName+".sock")
					Expect(fakeRemoteClientFactory.NewRemoteClientArgsForCall(0)).To(Equal(address))
				})
				It("should error for invalid sock endpoint address", func() {
					fakeRemoteClientFactory.NewRemoteClientReturns(nil, fmt.Errorf("invalid address"))
					_, err := driverFactory.DockerDriver(testLogger, driverName, defaultPluginsDirectory, driverName+".sock")
					Expect(err).To(HaveOccurred())
				})
			})
		}
	})

	Context("when valid driver spec is not discovered", func() {
		var (
			fakeRemoteClientFactory *voldriverfakes.FakeRemoteClientFactory
			fakeDriver              *voldriverfakes.FakeDriver
			driverFactory           voldiscoverers.DockerDriverFactory
		)
		BeforeEach(func() {
			driverName = "some-driver-name"
			fakeRemoteClientFactory = new(voldriverfakes.FakeRemoteClientFactory)
			fakeDriver = new(voldriverfakes.FakeDriver)
			fakeRemoteClientFactory.NewRemoteClientReturns(fakeDriver, nil)
			driverFactory = voldiscoverers.NewDockerDriverFactoryWithRemoteClientFactory(fakeRemoteClientFactory)

		})
		It("should error", func() {
			_, err := driverFactory.DockerDriver(testLogger, driverName, defaultPluginsDirectory, driverName+".spec")
			Expect(err).To(HaveOccurred())
		})
	})

})
