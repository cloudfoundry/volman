package voldocker_test

import (
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/voldriver"
	"code.cloudfoundry.org/voldriver/voldriverfakes"
	"code.cloudfoundry.org/volman/voldocker"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/volman"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("DockerDriverMounter", func() {
	var (
		volumeId      string
		logger        *lagertest.TestLogger
		dockerPlugin  volman.Plugin
		fakeVoldriver *voldriverfakes.FakeDriver
	)

	BeforeEach(func() {
		volumeId = "fake-volume"
		logger = lagertest.NewTestLogger("docker-mounter-test")
		fakeVoldriver = &voldriverfakes.FakeDriver{}
		dockerPlugin = voldocker.NewDockerPluginWithDriver(fakeVoldriver)
	})

	Describe("Mount", func() {
		Context("when given a driver", func() {

			Context("mount", func() {

				BeforeEach(func() {
					mountResponse := voldriver.MountResponse{Mountpoint: "/var/vcap/data/mounts/" + volumeId}
					fakeVoldriver.MountReturns(mountResponse)
				})

				It("should be able to mount without warning", func() {
					mountPath, err := dockerPlugin.Mount(logger, volumeId, map[string]interface{}{"volume_id": volumeId})
					Expect(err).NotTo(HaveOccurred())
					Expect(mountPath).NotTo(Equal(""))
					Expect(logger.Buffer()).NotTo(gbytes.Say("Invalid or dangerous mountpath"))
				})

				It("should not be able to mount if mount fails", func() {
					mountResponse := voldriver.MountResponse{Err: "an error"}
					fakeVoldriver.MountReturns(mountResponse)
					_, err := dockerPlugin.Mount(logger, volumeId, map[string]interface{}{"volume_id": volumeId})
					Expect(err).To(HaveOccurred())
				})

				Context("with bad mount path", func() {
					var err error
					BeforeEach(func() {
						mountResponse := voldriver.MountResponse{Mountpoint: "/var/tmp"}
						fakeVoldriver.MountReturns(mountResponse)
					})

					JustBeforeEach(func() {
						_, err = dockerPlugin.Mount(logger, volumeId, map[string]interface{}{"volume_id": volumeId})
					})

					It("should return a warning in the log", func() {
						Expect(err).NotTo(HaveOccurred())
						Expect(logger.Buffer()).To(gbytes.Say("Invalid or dangerous mountpath"))
					})
				})

				Context("with safe error", func() {
					var (
						err         error
						safeError   voldriver.SafeError
						unsafeError error
						errString   string
					)

					JustBeforeEach(func() {

						mountResponse := voldriver.MountResponse{Err: errString}
						fakeVoldriver.MountReturns(mountResponse)
						_, err = dockerPlugin.Mount(logger, volumeId, map[string]interface{}{"volume_id": volumeId})
					})

					Context("with safe error msg", func() {
						BeforeEach(func() {
							safeError = voldriver.SafeError{SafeDescription: "safe-badness"}
							errBytes, err := json.Marshal(safeError)
							Expect(err).NotTo(HaveOccurred())
							errString = string(errBytes[:])
						})

						It("should return a safe error", func() {
							Expect(err).To(HaveOccurred())
							_, ok := err.(voldriver.SafeError)
							Expect(ok).To(Equal(true))
							Expect(err.Error()).To(Equal("safe-badness"))
						})
					})

					Context("with unsafe error msg", func() {
						BeforeEach(func() {
							unsafeError = errors.New("unsafe-badness")
							errString = unsafeError.Error()
						})

						It("should return regular error", func() {
							Expect(err).To(HaveOccurred())
							_, ok := err.(voldriver.SafeError)
							Expect(ok).To(Equal(false))
							Expect(err.Error()).To(Equal("unsafe-badness"))
						})

					})

					Context("with a really unsafe error msg", func() {
						BeforeEach(func() {
							errString = "{ badness"
						})

						It("should return regular error", func() {
							Expect(err).To(HaveOccurred())
							_, ok := err.(voldriver.SafeError)
							Expect(ok).To(Equal(false))
							Expect(err.Error()).To(Equal("{ badness"))
						})
					})
				})

			})
		})
	})

	Describe("Unmount", func() {
		It("should be able to unmount", func() {
			err := dockerPlugin.Unmount(logger, volumeId)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeVoldriver.UnmountCallCount()).To(Equal(1))
			Expect(fakeVoldriver.RemoveCallCount()).To(Equal(0))
		})

		It("should not be able to unmount when driver unmount fails", func() {
			fakeVoldriver.UnmountReturns(voldriver.ErrorResponse{Err: "unmount failure"})
			err := dockerPlugin.Unmount(logger, volumeId)
			Expect(err).To(HaveOccurred())
		})

		Context("with safe error", func() {
			var (
				err         error
				safeError   voldriver.SafeError
				unsafeError error
				errString   string
			)

			JustBeforeEach(func() {
				fakeVoldriver.UnmountReturns(voldriver.ErrorResponse{Err: errString})
				err = dockerPlugin.Unmount(logger, volumeId)
			})

			Context("with safe error msg", func() {
				BeforeEach(func() {
					safeError = voldriver.SafeError{SafeDescription: "safe-badness"}
					errBytes, err := json.Marshal(safeError)
					Expect(err).NotTo(HaveOccurred())
					errString = string(errBytes[:])
				})

				It("should return a safe error", func() {
					Expect(err).To(HaveOccurred())
					_, ok := err.(voldriver.SafeError)
					Expect(ok).To(Equal(true))
					Expect(err.Error()).To(Equal("safe-badness"))
				})
			})

			Context("with unsafe error msg", func() {
				BeforeEach(func() {
					unsafeError = errors.New("unsafe-badness")
					errString = unsafeError.Error()
				})

				It("should return regular error", func() {
					Expect(err).To(HaveOccurred())
					_, ok := err.(voldriver.SafeError)
					Expect(ok).To(Equal(false))
					Expect(err.Error()).To(Equal("unsafe-badness"))
				})

			})

			Context("with a really unsafe error msg", func() {
				BeforeEach(func() {
					errString = "{ badness"
				})

				It("should return regular error", func() {
					Expect(err).To(HaveOccurred())
					_, ok := err.(voldriver.SafeError)
					Expect(ok).To(Equal(false))
					Expect(err.Error()).To(Equal("{ badness"))
				})
			})
		})

	})

	Describe("ListVolumes", func() {
		var (
			volumes []string
			err     error
		)
		BeforeEach(func() {
			listResponse := voldriver.ListResponse{Volumes: []voldriver.VolumeInfo{
				{Name: "fake_volume_1"},
				{Name: "fake_volume_2"},
			}}
			fakeVoldriver.ListReturns(listResponse)
		})

		JustBeforeEach(func() {
			volumes, err = dockerPlugin.ListVolumes(logger)
		})

		It("should be able list volumes", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(volumes).To(ContainElement("fake_volume_1"))
			Expect(volumes).To(ContainElement("fake_volume_2"))
		})

		Context("when the driver returns an err response", func() {
			BeforeEach(func() {
				listResponse := voldriver.ListResponse{Volumes: []voldriver.VolumeInfo{
					{Name: "fake_volume_1"},
					{Name: "fake_volume_2"},
				}, Err: "badness"}
				fakeVoldriver.ListReturns(listResponse)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})

})
