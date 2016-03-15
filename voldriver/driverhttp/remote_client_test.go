package driverhttp_test

import (
	"net/http"

	"github.com/cloudfoundry-incubator/volman/volmanfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/cloudfoundry-incubator/volman/voldriver/driverhttp"
	"github.com/pivotal-golang/lager/lagertest"

	"bytes"
	"fmt"
)

var _ = Describe("RemoteClient", func() {

	var (
		testLogger              = lagertest.NewTestLogger("FakeDriver Server Test")
		httpClient              *volmanfakes.FakeClient
		driver                  voldriver.Driver
		validHttpMountResponse  *http.Response
		validHttpCreateResponse *http.Response
		invalidHttpResponse     *http.Response
	)

	BeforeEach(func() {
		httpClient = new(volmanfakes.FakeClient)
		driver = driverhttp.NewRemoteClientWithHttpClient("http://127.0.0.1:8080", httpClient)
		validHttpCreateResponse = &http.Response{
			StatusCode: 200,
		}

		validHttpMountResponse = &http.Response{
			StatusCode: 200,
			Body:       stringCloser{bytes.NewBufferString("{\"Mountpoint\":\"somePath\"}")},
		}
	})

	Context("when the driver returns as error", func() {

		BeforeEach(func() {
			httpClient = new(volmanfakes.FakeClient)
			driver = driverhttp.NewRemoteClientWithHttpClient("http://127.0.0.1:8080", httpClient)
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

	Context("when the driver returns successful", func() {
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

			By("signaling an error")
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

	})

	Context("when the driver is malicious", func() {

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

	Context("when the http transport fails", func() {

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

	})

})
