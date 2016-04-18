package driverhttp_test

import (
	"net/http"
	"time"

	"bytes"
	"fmt"

	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/cloudfoundry-incubator/volman/voldriver/driverhttp"
	"github.com/cloudfoundry-incubator/volman/volmanfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("RemoteClient", func() {

	var (
		testLogger                = lagertest.NewTestLogger("FakeDriver Server Test")
		httpClient                *volmanfakes.FakeClient
		driver                    voldriver.Driver
		validHttpMountResponse    *http.Response
		validHttpCreateResponse   *http.Response
		validHttpActivateResponse *http.Response
		invalidHttpResponse       *http.Response
	)

	BeforeEach(func() {
		httpClient = new(volmanfakes.FakeClient)
		driver = driverhttp.NewRemoteClientWithClient("http://127.0.0.1:8080", httpClient)
		validHttpCreateResponse = &http.Response{
			StatusCode: 200,
		}

		validHttpMountResponse = &http.Response{
			StatusCode: 200,
			Body:       stringCloser{bytes.NewBufferString("{\"Mountpoint\":\"somePath\"}")},
		}

		validHttpActivateResponse = &http.Response{
			StatusCode: 200,
			Body:       stringCloser{bytes.NewBufferString("{\"Implements\":\"VolumeDriver\"}")},
		}
	})

	Context("when the driver returns as error and the transport is TCP", func() {

		BeforeEach(func() {
			httpClient = new(volmanfakes.FakeClient)
			driver = driverhttp.NewRemoteClientWithClient("http://127.0.0.1:8080", httpClient)
			invalidHttpResponse = &http.Response{
				StatusCode: 500,
				Body:       stringCloser{bytes.NewBufferString("{\"Err\":\"some error string\"}")},
			}
		})

		It("should not be able to mount", func() {

			httpClient.DoReturns(invalidHttpResponse, nil)

			volumeId := "fake-volume"
			mountResponse := driver.Mount(testLogger, voldriver.MountRequest{Name: volumeId})

			By("signaling an error")
			Expect(mountResponse.Err).To(Equal("some error string"))
			Expect(mountResponse.Mountpoint).To(Equal(""))
		})

		It("should not be able to unmount", func() {

			httpClient.DoReturns(invalidHttpResponse, nil)

			volumeId := "fake-volume"
			unmountResponse := driver.Unmount(testLogger, voldriver.UnmountRequest{Name: volumeId})

			By("signaling an error")
			Expect(unmountResponse.Err).To(Equal("some error string"))
		})

	})

	Context("when the driver returns successful and the transport is TCP", func() {
		var volumeId string

		BeforeEach(func() {
			httpClient.DoReturns(validHttpCreateResponse, nil)
			volumeId = "fake-volume"
			createResponse := driver.Create(testLogger, voldriver.CreateRequest{Name: volumeId, Opts: map[string]interface{}{"volume_id": volumeId}})
			Expect(createResponse.Err).To(Equal(""))
		})

		It("should be able to mount", func() {
			httpClient.DoReturns(validHttpMountResponse, nil)

			mountResponse := driver.Mount(testLogger, voldriver.MountRequest{Name: volumeId})

			By("giving back a path with no error")
			Expect(mountResponse.Err).To(Equal(""))
			Expect(mountResponse.Mountpoint).To(Equal("somePath"))
		})

		It("should be able to unmount", func() {

			validHttpUnmountResponse := &http.Response{
				StatusCode: 200,
			}

			httpClient.DoReturns(validHttpUnmountResponse, nil)

			volumeId := "fake-volume"
			unmountResponse := driver.Unmount(testLogger, voldriver.UnmountRequest{Name: volumeId})

			Expect(unmountResponse.Err).To(Equal(""))
		})

		It("should be able to activate", func() {
			httpClient.DoReturns(validHttpActivateResponse, nil)

			activateResponse := driver.Activate(testLogger)

			By("giving back a path with no error")
			Expect(activateResponse.Err).To(Equal(""))
			Expect(activateResponse.Implements).To(Equal("VolumeDriver"))
		})
	})

	Context("when the driver is malicious and the transport is TCP", func() {

		BeforeEach(func() {
		})

		It("should not be able to mount", func() {

			validHttpMountResponse = &http.Response{
				StatusCode: 200,
				Body:       stringCloser{bytes.NewBufferString("i am trying to pown your system")},
			}

			httpClient.DoReturns(validHttpMountResponse, nil)

			volumeId := "fake-volume"
			mountResponse := driver.Mount(testLogger, voldriver.MountRequest{Name: volumeId})

			By("signaling an error")
			Expect(mountResponse.Err).NotTo(Equal(""))
			Expect(mountResponse.Mountpoint).To(Equal(""))
		})

		It("should still not be able to mount", func() {

			invalidHttpResponse := &http.Response{
				StatusCode: 500,
				Body:       stringCloser{bytes.NewBufferString("i am trying to pown your system")},
			}

			httpClient.DoReturns(invalidHttpResponse, nil)

			volumeId := "fake-volume"
			mountResponse := driver.Mount(testLogger, voldriver.MountRequest{Name: volumeId})

			By("signaling an error")
			Expect(mountResponse.Err).NotTo(Equal(""))
			Expect(mountResponse.Mountpoint).To(Equal(""))
		})

		It("should not be able to unmount", func() {

			validHttpUnmountResponse := &http.Response{
				StatusCode: 500,
				Body:       stringCloser{bytes.NewBufferString("i am trying to pown your system")},
			}

			httpClient.DoReturns(validHttpUnmountResponse, nil)

			volumeId := "fake-volume"
			unmountResponse := driver.Unmount(testLogger, voldriver.UnmountRequest{Name: volumeId})

			Expect(unmountResponse.Err).NotTo(Equal(""))
		})

	})

	Context("when the http transport fails and the transport is TCP", func() {

		BeforeEach(func() {
		})

		It("should fail to mount", func() {

			httpClient.DoReturns(nil, fmt.Errorf("connection failed"))

			volumeId := "fake-volume"
			mountResponse := driver.Mount(testLogger, voldriver.MountRequest{Name: volumeId})

			By("signaling an error")
			Expect(mountResponse.Err).To(Equal("connection failed"))
		})

		It("should fail to unmount", func() {

			httpClient.DoReturns(nil, fmt.Errorf("connection failed"))

			volumeId := "fake-volume"
			unmountResponse := driver.Unmount(testLogger, voldriver.UnmountRequest{Name: volumeId})

			By("signaling an error")
			Expect(unmountResponse.Err).NotTo(Equal(""))
		})

		It("should still fail to unmount", func() {

			invalidHttpResponse := &http.Response{
				StatusCode: 500,
				Body:       errCloser{bytes.NewBufferString("")},
			}

			httpClient.DoReturns(invalidHttpResponse, nil)

			volumeId := "fake-volume"
			unmountResponse := driver.Unmount(testLogger, voldriver.UnmountRequest{Name: volumeId})

			Expect(unmountResponse.Err).NotTo(Equal(""))
		})

		It("should fail to activate", func() {
			httpClient.DoReturns(nil, fmt.Errorf("connection failed"))

			activateResponse := driver.Activate(testLogger)

			By("signaling an error")
			Expect(activateResponse.Err).NotTo(Equal(""))
		})

	})

	Context("when the transport is unix", func() {
		var volumeId string

		BeforeEach(func() {

			httpClient = new(volmanfakes.FakeClient)
			volumeId = "fake-volume"
			fakedriverUnixServerProcess = ginkgomon.Invoke(unixRunner)

			time.Sleep(time.Millisecond * 1000)

			driver = driverhttp.NewRemoteClientWithClient(socketPath, httpClient)
			validHttpMountResponse = &http.Response{
				StatusCode: 200,
				Body:       stringCloser{bytes.NewBufferString("{\"Mountpoint\":\"somePath\"}")},
			}

		})

		It("should be able to mount", func() {
			httpClient.DoReturns(validHttpMountResponse, nil)
			mountResponse := driver.Mount(testLogger, voldriver.MountRequest{Name: volumeId})

			By("returning a mountpoint without errors")
			Expect(mountResponse.Err).To(Equal(""))
			Expect(mountResponse.Mountpoint).To(Equal("somePath"))
		})

		It("should be able to unmount", func() {

			validHttpUnmountResponse := &http.Response{
				StatusCode: 200,
			}

			httpClient.DoReturns(validHttpUnmountResponse, nil)

			volumeId := "fake-volume"
			unmountResponse := driver.Unmount(testLogger, voldriver.UnmountRequest{Name: volumeId})

			Expect(unmountResponse.Err).To(Equal(""))
		})

		It("should be able to activate", func() {
			httpClient.DoReturns(validHttpActivateResponse, nil)

			activateResponse := driver.Activate(testLogger)

			By("giving back a path with no error")
			Expect(activateResponse.Err).To(Equal(""))
			Expect(activateResponse.Implements).To(Equal("VolumeDriver"))
		})

	})
})
