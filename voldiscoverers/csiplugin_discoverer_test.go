package voldiscoverers_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"google.golang.org/grpc"

	"golang.org/x/net/context"

	"code.cloudfoundry.org/csiplugin"
	"code.cloudfoundry.org/csishim/csi_fake"
	"code.cloudfoundry.org/goshims/filepathshim"
	"code.cloudfoundry.org/goshims/filepathshim/filepath_fake"
	"code.cloudfoundry.org/goshims/grpcshim/grpc_fake"
	"code.cloudfoundry.org/goshims/osshim/os_fake"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/volman"
	"code.cloudfoundry.org/volman/vollocal"
	"github.com/container-storage-interface/spec/lib/go/csi"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/csishim"
	"code.cloudfoundry.org/volman/voldiscoverers"
)

var _ = Describe("CSIPluginDiscoverer", func() {
	var (
		discoverer             volman.Discoverer
		registry               volman.PluginRegistry
		logger                 *lagertest.TestLogger
		firstPluginsDirectory  string
		secondPluginsDirectory string
		fakeFilePath           *filepath_fake.FakeFilepath
		fakeGrpc               *grpc_fake.FakeGrpc
		fakeCsi                *csi_fake.FakeCsi
		fakeOs                 *os_fake.FakeOs
		fakeNodePlugin         *csi_fake.FakeNodeClient
		fakeIdentityPlugin     *csi_fake.FakeIdentityClient
		pluginPaths            []string
		drivers                map[string]volman.Plugin
		err                    error
		volumesRootDir         string
	)

	BeforeEach(func() {
		firstPluginsDirectory, err = ioutil.TempDir(os.TempDir(), "one")
		secondPluginsDirectory, err = ioutil.TempDir(os.TempDir(), "two")
		Expect(err).ShouldNot(HaveOccurred())
		fakeFilePath = &filepath_fake.FakeFilepath{}
		fakeGrpc = &grpc_fake.FakeGrpc{}
		fakeCsi = &csi_fake.FakeCsi{}
		fakeOs = &os_fake.FakeOs{}
		fakeNodePlugin = &csi_fake.FakeNodeClient{}
		fakeIdentityPlugin = &csi_fake.FakeIdentityClient{}
		pluginPaths = []string{firstPluginsDirectory}
		volumesRootDir = "/var/vcap/data/mounts"

		logger = lagertest.NewTestLogger("csi-plugin-discoverer-test")
		registry = vollocal.NewPluginRegistry()
	})

	JustBeforeEach(func() {
		discoverer = voldiscoverers.NewCsiPluginDiscovererWithShims(logger, registry, pluginPaths, &filepathshim.FilepathShim{}, fakeGrpc, fakeCsi, fakeOs, volumesRootDir)
	})

	Describe("#Discover", func() {
		JustBeforeEach(func() {
			drivers, err = discoverer.Discover(logger)
			registry.Set(drivers)
		})

		Context("given a single plugin path", func() {

			Context("given there are no csi plugins", func() {
				It("should add nothing to the plugin registry", func() {
					Expect(len(drivers)).To(Equal(0))
				})
			})

			Context("given there is a csi plugin to be discovered", func() {
				var (
					fileName      string
					pluginName    string
					vendorVersion string
					address       string
				)

				BeforeEach(func() {
					fileName = fmt.Sprintf("csi-spec-file-%d", config.GinkgoConfig.ParallelNode)
					pluginName = fmt.Sprintf("csi-plugin-%d", config.GinkgoConfig.ParallelNode)
					vendorVersion = fmt.Sprintf("vendor-version-%d", config.GinkgoConfig.ParallelNode)
					address = "127.0.0.1:50051"
					spec := csiplugin.CsiPluginSpec{
						Name:    fileName,
						Address: address,
					}
					err := csiplugin.WriteSpec(logger, firstPluginsDirectory, spec)
					Expect(err).NotTo(HaveOccurred())

					fakeCsi.NewNodeClientReturns(fakeNodePlugin)
					fakeCsi.NewIdentityClientReturns(fakeIdentityPlugin)
					fakeIdentityPlugin.GetPluginInfoReturns(
						&csi.GetPluginInfoResponse{
							Name:          pluginName,
							VendorVersion: vendorVersion,
						}, nil)
				})

				Context("given the node is available", func() {
					var (
						conn *grpc_fake.FakeClientConn
					)
					BeforeEach(func() {
						conn = new(grpc_fake.FakeClientConn)
						fakeGrpc.DialReturns(conn, nil)
						fakeNodePlugin.NodeProbeReturns(&csi.NodeProbeResponse{}, nil)
					})

					Context("when the node supports our version of CSI", func() {
						BeforeEach(func() {
							fakeIdentityPlugin.GetSupportedVersionsReturns(&csi.GetSupportedVersionsResponse{
								SupportedVersions: []*csi.Version{&csishim.CsiVersion},
							}, nil)
						})

						It("should discover it and add it to the plugin registry", func() {
							Expect(fakeGrpc.DialCallCount()).To(Equal(1))
							actualAddr, _ := fakeGrpc.DialArgsForCall(0)
							Expect(actualAddr).To(Equal(address))

							Expect(fakeIdentityPlugin.GetSupportedVersionsCallCount()).To(BeNumerically(">", 0))
							Expect(fakeIdentityPlugin.GetPluginInfoCallCount()).To(BeNumerically(">", 0))

							Expect(fakeNodePlugin.NodeProbeCallCount()).To(Equal(1))

							Expect(len(drivers)).To(Equal(1))

							_, pluginFound := drivers[pluginName]
							Expect(pluginFound).To(Equal(true))

							Expect(conn.CloseCallCount()).To(Equal(1))
						})

						Context("given re-discovery", func() {
							JustBeforeEach(func() {
								drivers, err = discoverer.Discover(logger)
								registry.Set(drivers)
							})

							It("should not update the plugin registry", func() {
								Expect(drivers).To(HaveLen(1))
								Expect(fakeCsi.NewNodeClientCallCount()).To(Equal(1))
							})
						})

						Context("when the plugin's spec is changed", func() {
							var (
								updatedAddress string
							)
							JustBeforeEach(func() {
								updatedAddress = "127.0.0.1:99999"
								spec := csiplugin.CsiPluginSpec{
									Name:    fileName,
									Address: updatedAddress,
								}
								err := csiplugin.WriteSpec(logger, firstPluginsDirectory, spec)
								Expect(err).NotTo(HaveOccurred())

								drivers, err = discoverer.Discover(logger)
								Expect(drivers).To(HaveLen(1))
								registry.Set(drivers)
							})

							It("should re-discover the plugin and update the registry", func() {
								Expect(fakeGrpc.DialCallCount()).To(Equal(2))
								actualAddr1, _ := fakeGrpc.DialArgsForCall(0)
								Expect(actualAddr1).To(Equal(address))
								actualAddr2, _ := fakeGrpc.DialArgsForCall(1)
								Expect(actualAddr2).To(Equal(updatedAddress))

								Expect(fakeCsi.NewNodeClientCallCount()).To(Equal(2))
								Expect(fakeNodePlugin.NodeProbeCallCount()).To(Equal(2))

								Expect(conn.CloseCallCount()).To(Equal(2))

								Expect(len(drivers)).To(Equal(1))

								_, pluginFound := drivers[pluginName]
								Expect(pluginFound).To(Equal(true))
							})
						})
					})

					Context("when the node does not support our version of CSI", func() {
						BeforeEach(func() {
							fakeIdentityPlugin.GetSupportedVersionsReturns(&csi.GetSupportedVersionsResponse{
								SupportedVersions: []*csi.Version{{Major: 9, Minor: 9, Patch: 9}},
							}, nil)
						})

						It("should not discover the plugin", func() {
							Expect(fakeGrpc.DialCallCount()).To(Equal(1))
							actualAddr, _ := fakeGrpc.DialArgsForCall(0)
							Expect(actualAddr).To(Equal(address))

							Expect(fakeIdentityPlugin.GetSupportedVersionsCallCount()).To(BeNumerically(">", 0))
							Expect(fakeIdentityPlugin.GetPluginInfoCallCount()).To(Equal(0))

							Expect(fakeNodePlugin.NodeProbeCallCount()).To(Equal(0))
						})
					})
				})

				Context("given the node is not available", func() {
					var (
						conn *grpc_fake.FakeClientConn
					)
					BeforeEach(func() {
						conn = new(grpc_fake.FakeClientConn)
						fakeGrpc.DialReturns(conn, nil)

						fakeCsi.NewIdentityClientReturns(fakeIdentityPlugin)
						fakeIdentityPlugin.GetSupportedVersionsReturns(nil, errors.New("connection-refused"))
					})

					It("should have discover it but not add it to the plugin registry", func() {
						Expect(fakeGrpc.DialCallCount()).To(Equal(1))
						actualAddr, _ := fakeGrpc.DialArgsForCall(0)
						Expect(actualAddr).To(Equal(address))

						Expect(fakeIdentityPlugin.GetSupportedVersionsCallCount()).To(Equal(1))
						Expect(fakeIdentityPlugin.GetPluginInfoCallCount()).To(Equal(0))
						Expect(fakeNodePlugin.NodeProbeCallCount()).To(Equal(0))

						Expect(conn.CloseCallCount()).To(Equal(1))

						Expect(len(drivers)).To(Equal(0))
					})
				})
			})
		})

		Context("given more than one plugin path", func() {
			var (
				conn *grpc_fake.FakeClientConn
			)

			BeforeEach(func() {
				pluginPaths = []string{firstPluginsDirectory, secondPluginsDirectory}
			})

			Context("given multiple plugins to be discovered, in multiple directories", func() {
				var (
					specFileName  string
					pluginName    string
					address       string
					spec          csiplugin.CsiPluginSpec
					specFileName2 string
					pluginName2   string
					address2      string
					spec2         csiplugin.CsiPluginSpec
					err           error
				)

				BeforeEach(func() {
					specFileName = fmt.Sprintf("csi-spec-file-%d", config.GinkgoConfig.ParallelNode)
					pluginName = fmt.Sprintf("csi-plugin-%d", config.GinkgoConfig.ParallelNode)
					address = "127.0.0.1:50051"
					spec = csiplugin.CsiPluginSpec{
						Name:    specFileName,
						Address: address,
					}
					err = csiplugin.WriteSpec(logger, firstPluginsDirectory, spec)
					Expect(err).NotTo(HaveOccurred())

					specFileName2 = fmt.Sprintf("csi-spec-file-2-%d", config.GinkgoConfig.ParallelNode)
					pluginName2 = fmt.Sprintf("csi-plugin-2-%d", config.GinkgoConfig.ParallelNode)
					address2 = "127.0.0.1:50061"
					spec2 = csiplugin.CsiPluginSpec{
						Name:    specFileName2,
						Address: address2,
					}
					err = csiplugin.WriteSpec(logger, secondPluginsDirectory, spec2)
					Expect(err).NotTo(HaveOccurred())

					conn = new(grpc_fake.FakeClientConn)
					fakeGrpc.DialReturns(conn, nil)

					// make both plugins active
					fakeCsi.NewNodeClientReturns(fakeNodePlugin)
					fakeNodePlugin.NodeProbeReturns(&csi.NodeProbeResponse{}, nil)

					fakeCsi.NewIdentityClientReturns(fakeIdentityPlugin)
					fakeIdentityPlugin.GetSupportedVersionsReturns(&csi.GetSupportedVersionsResponse{
						SupportedVersions: []*csi.Version{&csishim.CsiVersion},
					}, nil)

					i := 0
					fakeIdentityPlugin.GetPluginInfoStub = func(ctx context.Context, req *csi.GetPluginInfoRequest, options ...grpc.CallOption) (*csi.GetPluginInfoResponse, error) {
						if i == 0 {
							i++
							return &csi.GetPluginInfoResponse{
								Name:          pluginName,
								VendorVersion: "9.9.9",
							}, nil

						}
						return &csi.GetPluginInfoResponse{
							Name:          pluginName2,
							VendorVersion: "9.9.9",
						}, nil
					}
				})

				It("should discover both plugins", func() {
					Expect(len(drivers)).To(Equal(2))

					_, pluginFound := drivers[pluginName]
					Expect(pluginFound).To(Equal(true))

					_, plugin2Found := drivers[pluginName2]
					Expect(plugin2Found).To(Equal(true))
				})
			})

			Context("given the same plugin in multiple directories", func() {
				var (
					pluginName    string
					vendorVersion string
					address       string
					spec          csiplugin.CsiPluginSpec
					fileName      string
					err           error
				)

				BeforeEach(func() {
					fileName = fmt.Sprintf("csi-file-name-%d", config.GinkgoConfig.ParallelNode)
					pluginName = fmt.Sprintf("csi-plugin-%d", config.GinkgoConfig.ParallelNode)
					vendorVersion = fmt.Sprintf("csi-vendor-%d", config.GinkgoConfig.ParallelNode)
					address = "127.0.0.1:50051"
					spec = csiplugin.CsiPluginSpec{
						Name:    fileName,
						Address: address,
					}
					err = csiplugin.WriteSpec(logger, firstPluginsDirectory, spec)
					Expect(err).NotTo(HaveOccurred())

					fmt.Sprintf("csi-plugin-%d", config.GinkgoConfig.ParallelNode)
					err = csiplugin.WriteSpec(logger, secondPluginsDirectory, spec)
					Expect(err).NotTo(HaveOccurred())

					conn = new(grpc_fake.FakeClientConn)
					fakeGrpc.DialReturns(conn, nil)

					// make both plugins active
					fakeCsi.NewNodeClientReturns(fakeNodePlugin)
					fakeNodePlugin.NodeProbeReturns(&csi.NodeProbeResponse{}, nil)

					fakeCsi.NewIdentityClientReturns(fakeIdentityPlugin)
					fakeCsi.NewIdentityClientReturns(fakeIdentityPlugin)
					fakeIdentityPlugin.GetSupportedVersionsReturns(&csi.GetSupportedVersionsResponse{
						SupportedVersions: []*csi.Version{&csishim.CsiVersion},
					}, nil)

					fakeIdentityPlugin.GetPluginInfoReturns(
						&csi.GetPluginInfoResponse{
							Name:          pluginName,
							VendorVersion: vendorVersion,
						}, nil)

				})

				It("should discover the plugin and add it to the registry once only", func() {
					Expect(len(drivers)).To(Equal(1))

					_, pluginFound := drivers[pluginName]
					Expect(pluginFound).To(Equal(true))
				})
			})
		})
	})
})
