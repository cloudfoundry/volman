package fakedriver_test

import (
	"errors"
	"fmt"
	"os"

	"github.com/cloudfoundry-incubator/volman/fakedriver"
	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/cloudfoundry-incubator/volman/volmanfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("Local Driver", func() {
	var logger lager.Logger
	var fakeFileSystem *volmanfakes.FakeFileSystem
	var localDriver *fakedriver.LocalDriver

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("fakedriver-local")

		fakeFileSystem = &volmanfakes.FakeFileSystem{}
		localDriver = fakedriver.NewLocalDriver(fakeFileSystem)
	})

	Describe("Mount", func() {
		Context("when the volume has been created", func() {
			const volumeName = "test-volume-name"

			BeforeEach(func() {
				createSuccessful(logger, localDriver, volumeName)
				mountSuccessful(logger, localDriver, volumeName, fakeFileSystem)
			})

			It("mounts the volume on the local filesystem", func() {
				Expect(fakeFileSystem.AbsCallCount()).To(Equal(1))
				Expect(fakeFileSystem.MkdirAllCallCount()).To(Equal(1))
				createdDir, permissions := fakeFileSystem.MkdirAllArgsForCall(0)
				Expect(createdDir).To(Equal("/some/temp/dir/_volumes/test-volume-id"))
				Expect(permissions).To(BeEquivalentTo(0777))
			})

			It("returns the mount point on a /VolumeDriver.Get response", func() {
				getResponse := getSuccessful(logger, localDriver, volumeName)
				Expect(getResponse.Volume.Mountpoint).To(Equal("/some/temp/dir/_volumes/test-volume-id"))
			})
		})

		Context("when the volume has not been created", func() {
			It("returns an error", func() {
				mountResponse := localDriver.Mount(logger, voldriver.MountRequest{
					Name: "bla",
				})
				Expect(mountResponse.Err).To(Equal("Volume 'bla' must be created before being mounted"))
			})
		})
	})

	Describe("Unmount", func() {
		const volumeName = "volumeName"

		Context("when a volume has been created", func() {
			BeforeEach(func() {
				createSuccessful(logger, localDriver, volumeName)
			})

			Context("when a volume has been mounted", func() {
				BeforeEach(func() {
					mountSuccessful(logger, localDriver, volumeName, fakeFileSystem)
				})

				It("/VolumeDriver.Get returns no mountpoint", func() {
					unmountSuccessful(logger, localDriver, volumeName)
					getResponse := getSuccessful(logger, localDriver, volumeName)
					Expect(getResponse.Volume.Mountpoint).To(Equal(""))
				})

				It("/VolumeDriver.Unmount removes mountpath from OS", func() {
					unmountSuccessful(logger, localDriver, volumeName)
					Expect(fakeFileSystem.RemoveAllCallCount()).To(Equal(1))
				})

				Context("when the mountpath is not found on the filesystem", func() {
					var unmountResponse voldriver.ErrorResponse

					BeforeEach(func() {
						fakeFileSystem.StatReturns(nil, os.ErrNotExist)
						unmountResponse = localDriver.Unmount(logger, voldriver.UnmountRequest{
							Name: volumeName,
						})
					})

					It("returns an error", func() {
						Expect(unmountResponse.Err).To(Equal(fmt.Sprintf("Volume %s does not exist, nothing to do!", volumeName)))
					})

					It("/VolumeDriver.Get still returns the mountpoint", func() {
						getResponse := getSuccessful(logger, localDriver, volumeName)
						Expect(getResponse.Volume.Mountpoint).NotTo(Equal(""))
					})
				})

				Context("when the mountpath cannot be accessed", func() {
					var unmountResponse voldriver.ErrorResponse

					BeforeEach(func() {
						fakeFileSystem.StatReturns(nil, errors.New("something weird"))
						unmountResponse = localDriver.Unmount(logger, voldriver.UnmountRequest{
							Name: volumeName,
						})
					})

					It("returns an error", func() {
						Expect(unmountResponse.Err).To(Equal("Error establishing whether volume exists"))
					})

					It("/VolumeDriver.Get still returns the mountpoint", func() {
						getResponse := getSuccessful(logger, localDriver, volumeName)
						Expect(getResponse.Volume.Mountpoint).NotTo(Equal(""))
					})
				})

				Context("when removing the mountpath errors", func() {
					var unmountResponse voldriver.ErrorResponse

					BeforeEach(func() {
						fakeFileSystem.RemoveAllReturns(errors.New("an error"))

						unmountResponse = localDriver.Unmount(logger, voldriver.UnmountRequest{
							Name: volumeName,
						})
					})

					It("returns an error", func() {
						Expect(unmountResponse.Err).To(Equal("Failed removing mount path: an error"))
					})

					It("/VolumeDriver.Get still returns the mountpoint", func() {
						getResponse := getSuccessful(logger, localDriver, volumeName)
						Expect(getResponse.Volume.Mountpoint).NotTo(Equal(""))
					})
				})
			})

			Context("when the volume has not been mounted", func() {
				It("returns an error", func() {
					unmountResponse := localDriver.Unmount(logger, voldriver.UnmountRequest{
						Name: volumeName,
					})

					Expect(unmountResponse.Err).To(Equal("Volume not previously mounted"))
				})
			})
		})

		Context("when the volume has not been created", func() {
			It("returns an error", func() {
				unmountResponse := localDriver.Unmount(logger, voldriver.UnmountRequest{
					Name: volumeName,
				})

				Expect(unmountResponse.Err).To(Equal(fmt.Sprintf("Volume '%s' not found", volumeName)))
			})
		})
	})

	Describe("Create", func() {
		Context("when a volume ID is not provided", func() {
			It("returns an error", func() {
				createResponse := localDriver.Create(logger, voldriver.CreateRequest{
					Name: "volume",
					Opts: map[string]interface{}{
						"nonsense": "bla",
					},
				})

				Expect(createResponse.Err).To(Equal("Missing mandatory 'volume_id' field in 'Opts'"))
			})
		})

		Context("when a second create is called with the same volume ID", func() {
			BeforeEach(func() {
				createSuccessful(logger, localDriver, "volume")
			})

			Context("with the same opts", func() {
				It("does nothing", func() {
					createSuccessful(logger, localDriver, "volume")
				})
			})

			Context("with a different opts", func() {
				It("returns an error", func() {
					createResponse := localDriver.Create(logger, voldriver.CreateRequest{
						Name: "volume",
						Opts: map[string]interface{}{
							"volume_id": "something_different_than_test",
						},
					})

					Expect(createResponse.Err).To(Equal("Volume 'volume' already exists with a different volume ID"))
				})
			})
		})
	})

	Describe("Get", func() {
		Context("when the volume has been created", func() {
			It("returns the volume name", func() {
				volumeName := "test-volume"
				createSuccessful(logger, localDriver, volumeName)
				getSuccessful(logger, localDriver, volumeName)
			})
		})

		Context("when the volume has not been created", func() {
			It("returns an error", func() {
				volumeName := "test-volume"
				getUnsuccessful(logger, localDriver, volumeName)
			})
		})
	})

	Describe("Remove", func() {
		const volumeName = "test-volume"
		Context("when the volume has been created", func() {
			BeforeEach(func() {
				createSuccessful(logger, localDriver, volumeName)
			})

			It("destroys volume", func() {
				removeResponse := localDriver.Remove(logger, voldriver.RemoveRequest{
					Name: volumeName,
				})
				Expect(removeResponse.Err).To(Equal(""))
				getUnsuccessful(logger, localDriver, volumeName)
			})
			Context("when volume has been mounted", func() {
				It("unmounts and destroys volume", func() {
					mountSuccessful(logger, localDriver, volumeName, fakeFileSystem)

					removeResponse := localDriver.Remove(logger, voldriver.RemoveRequest{
						Name: volumeName,
					})
					Expect(removeResponse.Err).To(Equal(""))
					getUnsuccessful(logger, localDriver, volumeName)
				})
			})
		})

		Context("when the volume has not been created", func() {
			It("returns an error", func() {
				removeResponse := localDriver.Remove(logger, voldriver.RemoveRequest{
					Name: volumeName,
				})
				Expect(removeResponse.Err).To(Equal("Volume test-volume not found"))
			})
		})
	})
})

func getUnsuccessful(logger lager.Logger, localDriver voldriver.Driver, volumeName string){
	getResponse := localDriver.Get(logger, voldriver.GetRequest{
		Name: volumeName,
	})

	Expect(getResponse.Err).To(Equal("Volume not found"))
	Expect(getResponse.Volume.Name).To(Equal(""))
}

func getSuccessful(logger lager.Logger, localDriver voldriver.Driver, volumeName string) voldriver.GetResponse {
	getResponse := localDriver.Get(logger, voldriver.GetRequest{
		Name: volumeName,
	})

	Expect(getResponse.Err).To(Equal(""))
	Expect(getResponse.Volume.Name).To(Equal(volumeName))
	return getResponse
}

func createSuccessful(logger lager.Logger, localDriver voldriver.Driver, volumeName string){
	createResponse := localDriver.Create(logger, voldriver.CreateRequest{
		Name: volumeName,
		Opts: map[string]interface{}{
			"volume_id": "test-volume-id",
		},
	})
	Expect(createResponse.Err).To(Equal(""))
}

func mountSuccessful(logger lager.Logger, localDriver voldriver.Driver, volumeName string, fakeFileSystem *volmanfakes.FakeFileSystem){
	fakeFileSystem.AbsReturns("/some/temp/dir", nil)
	mountResponse := localDriver.Mount(logger, voldriver.MountRequest{
	Name: volumeName,
	})
	Expect(mountResponse.Err).To(Equal(""))
	Expect(mountResponse.Mountpoint).To(Equal("/some/temp/dir/_volumes/test-volume-id"))
}

func unmountSuccessful(logger lager.Logger, localDriver voldriver.Driver, volumeName string) {
	unmountResponse := localDriver.Unmount(logger, voldriver.UnmountRequest{
		Name: volumeName,
	})
	Expect(unmountResponse.Err).To(Equal(""))
}
