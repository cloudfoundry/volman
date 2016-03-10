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
		testLogger             = lagertest.NewTestLogger("FakeDriver Server Test")
		httpClient             *volmanfakes.FakeClient
		driver                 voldriver.Driver
		validHttpMountResponse *http.Response
		invalidHttpResponse    *http.Response
	)

	BeforeEach(func() {
		httpClient = new(volmanfakes.FakeClient)
		driver = driverhttp.NewRemoteClientWithHttpClient("http://127.0.0.1:8080", httpClient)
		validHttpMountResponse = &http.Response{
			StatusCode: 200,
			Body:       stringCloser{bytes.NewBufferString("{\"Path\":\"somePath\"}")},
		}
	})

	Context("when the driver returns as error", func() {

		BeforeEach(func() {
			httpClient = new(volmanfakes.FakeClient)
			driver = driverhttp.NewRemoteClientWithHttpClient("http://127.0.0.1:8080", httpClient)
			invalidHttpResponse = &http.Response{
				StatusCode: 500,
				Body:       stringCloser{bytes.NewBufferString("{\"description\":\"some error string\"}")},
			}
		})

		It("should not be able to mount", func() {

			httpClient.DoReturns(invalidHttpResponse, nil)

			volumeId := "fake-volume"
			config := "Here is some config!"
			mountResponse, err := driver.Mount(testLogger, voldriver.MountRequest{VolumeId: volumeId, Config: config})

			By("signaling an error")
			Expect(err).To(HaveOccurred())
			Expect(mountResponse.Path).To(Equal(""))
		})

		It("should not be able to unmount", func() {

			httpClient.DoReturns(invalidHttpResponse, nil)

			volumeId := "fake-volume"
			err := driver.Unmount(testLogger, voldriver.UnmountRequest{VolumeId: volumeId})

			By("signaling an error")
			Expect(err).To(HaveOccurred())
		})

	})

	Context("when the driver returns successful", func() {

		BeforeEach(func() {
		})

		It("should be able to mount", func() {

			httpClient.DoReturns(validHttpMountResponse, nil)

			volumeId := "fake-volume"
			config := "Here is some config!"
			mountResponse, err := driver.Mount(testLogger, voldriver.MountRequest{VolumeId: volumeId, Config: config})

			By("signaling an error")
			Expect(err).NotTo(HaveOccurred())
			Expect(mountResponse.Path).To(Equal("somePath"))
		})

		It("should be able to unmount", func() {

			validHttpUnmountResponse := &http.Response{
				StatusCode: 200,
			}

			httpClient.DoReturns(validHttpUnmountResponse, nil)

			volumeId := "fake-volume"
			err := driver.Unmount(testLogger, voldriver.UnmountRequest{VolumeId: volumeId})

			Expect(err).NotTo(HaveOccurred())
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
			config := "Here is some config!"
			mountResponse, err := driver.Mount(testLogger, voldriver.MountRequest{VolumeId: volumeId, Config: config})

			By("signaling an error")
			Expect(err).To(HaveOccurred())
			Expect(mountResponse.Path).To(Equal(""))
		})

		It("should still not be able to mount", func() {

			invalidHttpResponse := &http.Response{
				StatusCode: 500,
				Body:       stringCloser{bytes.NewBufferString("i am trying to pown your system")},
			}

			httpClient.DoReturns(invalidHttpResponse, nil)

			volumeId := "fake-volume"
			config := "Here is some config!"
			mountResponse, err := driver.Mount(testLogger, voldriver.MountRequest{VolumeId: volumeId, Config: config})

			By("signaling an error")
			Expect(err).To(HaveOccurred())
			Expect(mountResponse.Path).To(Equal(""))
		})

		It("should not be able to unmount", func() {

			validHttpUnmountResponse := &http.Response{
				StatusCode: 500,
				Body:       stringCloser{bytes.NewBufferString("i am trying to pown your system")},
			}

			httpClient.DoReturns(validHttpUnmountResponse, nil)

			volumeId := "fake-volume"
			err := driver.Unmount(testLogger, voldriver.UnmountRequest{VolumeId: volumeId})

			Expect(err).To(HaveOccurred())
		})

	})

	Context("when the http transport fails", func() {

		BeforeEach(func() {
		})

		It("should fail to mount", func() {

			httpClient.DoReturns(nil, fmt.Errorf("connection failed"))

			volumeId := "fake-volume"
			config := "Here is some config!"
			_, err := driver.Mount(testLogger, voldriver.MountRequest{VolumeId: volumeId, Config: config})

			By("signaling an error")
			Expect(err).To(HaveOccurred())
		})

		It("should fail to unmount", func() {

			httpClient.DoReturns(nil, fmt.Errorf("connection failed"))

			volumeId := "fake-volume"
			err := driver.Unmount(testLogger, voldriver.UnmountRequest{VolumeId: volumeId})

			By("signaling an error")
			Expect(err).To(HaveOccurred())
		})

		It("should still fail to unmount", func() {

			invalidHttpResponse := &http.Response{
				StatusCode: 500,
				Body:       errCloser{bytes.NewBufferString("")},
			}

			httpClient.DoReturns(invalidHttpResponse, nil)

			volumeId := "fake-volume"
			err := driver.Unmount(testLogger, voldriver.UnmountRequest{VolumeId: volumeId})

			Expect(err).To(HaveOccurred())
		})

	})

})
