package driverhttp_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"fmt"

	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/cloudfoundry-incubator/volman/voldriver/driverhttp"
	"github.com/cloudfoundry-incubator/volman/volmanfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("Volman Driver Handlers", func() {

	Context("when generating http handlers", func() {
		var testLogger = lagertest.NewTestLogger("HandlersTest")

		It("should produce a handler with an activate route", func() {
			By("faking out the driver")
			driver := &volmanfakes.FakeDriver{}
			driver.ActivateReturns(voldriver.ActivateResponse{Implements: []string{"VolumeDriver"}})
			handler, err := driverhttp.NewHandler(testLogger, driver)
			Expect(err).NotTo(HaveOccurred())

			By("then fake serving the response using the handler")
			route, found := voldriver.Routes.FindRouteByName(voldriver.ActivateRoute)
			Expect(found).To(BeTrue())

			path := fmt.Sprintf("http://0.0.0.0%s", route.Path)
			httpRequest, err := http.NewRequest("POST", path, bytes.NewReader([]byte{}))
			Expect(err).NotTo(HaveOccurred())

			httpResponseRecorder := httptest.NewRecorder()
			handler.ServeHTTP(httpResponseRecorder, httpRequest)

			By("then deserialing the HTTP response")
			activateResponse := voldriver.ActivateResponse{}
			body, err := ioutil.ReadAll(httpResponseRecorder.Body)
			err = json.Unmarshal(body, &activateResponse)

			By("then expecting correct JSON conversion")
			Expect(err).ToNot(HaveOccurred())
			Expect(activateResponse.Implements).Should(Equal([]string{"VolumeDriver"}))
		})

		It("should produce a handler with a list route", func() {
			By("faking out the driver")
			driver := &volmanfakes.FakeDriver{}
			volume := voldriver.VolumeInfo{
				Name:       "fake-volume",
				Mountpoint: "fake-mountpoint",
			}
			listResponse := voldriver.ListResponse{
				Volumes: []voldriver.VolumeInfo{volume},
				Err:     "",
			}

			driver.ListReturns(listResponse)
			handler, err := driverhttp.NewHandler(testLogger, driver)
			Expect(err).NotTo(HaveOccurred())

			By("then fake serving the response using the handler")
			route, found := voldriver.Routes.FindRouteByName(voldriver.ListRoute)
			Expect(found).To(BeTrue())

			path := fmt.Sprintf("http://0.0.0.0%s", route.Path)
			httpRequest, err := http.NewRequest("POST", path, bytes.NewReader([]byte{}))
			Expect(err).NotTo(HaveOccurred())

			httpResponseRecorder := httptest.NewRecorder()
			handler.ServeHTTP(httpResponseRecorder, httpRequest)

			By("then deserialing the HTTP response")
			listResponse = voldriver.ListResponse{}
			body, err := ioutil.ReadAll(httpResponseRecorder.Body)
			err = json.Unmarshal(body, &listResponse)

			By("then expecting correct JSON conversion")
			Expect(err).ToNot(HaveOccurred())
			Expect(listResponse.Volumes[0].Name).Should(Equal("fake-volume"))
		})

		It("should produce a handler with a mount route", func() {
			By("faking out the driver")
			driver := &volmanfakes.FakeDriver{}
			driver.MountReturns(voldriver.MountResponse{Mountpoint: "dummy_path"})
			handler, err := driverhttp.NewHandler(testLogger, driver)
			Expect(err).NotTo(HaveOccurred())

			httpResponseRecorder := httptest.NewRecorder()
			MountRequest := voldriver.MountRequest{}
			mountJSONRequest, err := json.Marshal(MountRequest)
			Expect(err).NotTo(HaveOccurred())

			By("then fake serving the response using the handler")
			route, found := voldriver.Routes.FindRouteByName(voldriver.MountRoute)
			Expect(found).To(BeTrue())

			path := fmt.Sprintf("http://0.0.0.0%s", route.Path)
			httpRequest, err := http.NewRequest("POST", path, bytes.NewReader(mountJSONRequest))
			Expect(err).NotTo(HaveOccurred())

			testLogger.Info(fmt.Sprintf("%#v", httpResponseRecorder.Body))
			handler.ServeHTTP(httpResponseRecorder, httpRequest)

			By("then deserialing the HTTP response")
			mountResponse := voldriver.MountResponse{}
			body, err := ioutil.ReadAll(httpResponseRecorder.Body)
			err = json.Unmarshal(body, &mountResponse)

			By("then expecting correct JSON conversion")
			Expect(err).ToNot(HaveOccurred())
			Expect(mountResponse.Mountpoint).Should(Equal("dummy_path"))
		})

		It("should produce a handler with an unmount route", func() {
			By("faking out the driver")
			driver := &volmanfakes.FakeDriver{}
			driver.UnmountReturns(voldriver.ErrorResponse{})

			handler, err := driverhttp.NewHandler(testLogger, driver)
			Expect(err).NotTo(HaveOccurred())

			httpResponseRecorder := httptest.NewRecorder()
			unmountRequest := voldriver.UnmountRequest{}
			unmountJSONRequest, err := json.Marshal(unmountRequest)
			Expect(err).NotTo(HaveOccurred())

			By("then fake serving the response using the handler")
			route, found := voldriver.Routes.FindRouteByName(voldriver.UnmountRoute)
			Expect(found).To(BeTrue())

			path := fmt.Sprintf("http://0.0.0.0%s", route.Path)
			httpRequest, err := http.NewRequest("POST", path, bytes.NewReader(unmountJSONRequest))
			Expect(err).NotTo(HaveOccurred())
			handler.ServeHTTP(httpResponseRecorder, httpRequest)

			By("then expecting correct HTTP status code")
			Expect(httpResponseRecorder.Code).To(Equal(200))
		})

		It("should produce a handler with a get route", func() {
			By("faking out the driver")
			driver := &volmanfakes.FakeDriver{}
			driver.GetReturns(voldriver.GetResponse{Volume: voldriver.VolumeInfo{Name: "some-volume", Mountpoint: "dummy_path"}})
			handler, err := driverhttp.NewHandler(testLogger, driver)
			Expect(err).NotTo(HaveOccurred())

			httpResponseRecorder := httptest.NewRecorder()
			getRequest := voldriver.GetRequest{}
			getJSONRequest, err := json.Marshal(getRequest)
			Expect(err).NotTo(HaveOccurred())

			By("then fake serving the response using the handler")
			route, found := voldriver.Routes.FindRouteByName(voldriver.GetRoute)
			Expect(found).To(BeTrue())

			path := fmt.Sprintf("http://0.0.0.0%s", route.Path)
			httpRequest, err := http.NewRequest("POST", path, bytes.NewReader(getJSONRequest))
			Expect(err).NotTo(HaveOccurred())
			handler.ServeHTTP(httpResponseRecorder, httpRequest)

			By("then expecting correct HTTP status code")
			Expect(httpResponseRecorder.Code).To(Equal(200))
		})

		It("should produce a handler with a path route", func() {
			By("faking out the driver")
			driver := &volmanfakes.FakeDriver{}
			driver.PathReturns(voldriver.PathResponse{})
			handler, err := driverhttp.NewHandler(testLogger, driver)
			Expect(err).NotTo(HaveOccurred())

			httpResponseRecorder := httptest.NewRecorder()
			pathRequest := voldriver.PathRequest{Name: "some-volume"}
			pathJSONRequest, err := json.Marshal(pathRequest)
			Expect(err).NotTo(HaveOccurred())

			By("then fake serving the response using the handler")
			route, found := voldriver.Routes.FindRouteByName(voldriver.PathRoute)
			Expect(found).To(BeTrue())

			path := fmt.Sprintf("http://0.0.0.0%s", route.Path)
			httpRequest, err := http.NewRequest("POST", path, bytes.NewReader(pathJSONRequest))
			Expect(err).NotTo(HaveOccurred())
			handler.ServeHTTP(httpResponseRecorder, httpRequest)

			By("then expecting correct HTTP status code")
			Expect(httpResponseRecorder.Code).To(Equal(200))
		})

		It("should produce a handler with a create route", func() {
			By("faking out the driver")
			driver := &volmanfakes.FakeDriver{}
			driver.CreateReturns(voldriver.ErrorResponse{})
			handler, err := driverhttp.NewHandler(testLogger, driver)
			Expect(err).NotTo(HaveOccurred())

			httpResponseRecorder := httptest.NewRecorder()
			createRequest := voldriver.CreateRequest{Name: "some-volume"}
			createJSONRequest, err := json.Marshal(createRequest)
			Expect(err).NotTo(HaveOccurred())

			By("then fake serving the response using the handler")
			route, found := voldriver.Routes.FindRouteByName(voldriver.CreateRoute)
			Expect(found).To(BeTrue())

			path := fmt.Sprintf("http://0.0.0.0%s", route.Path)
			httpRequest, err := http.NewRequest("POST", path, bytes.NewReader(createJSONRequest))
			Expect(err).NotTo(HaveOccurred())
			handler.ServeHTTP(httpResponseRecorder, httpRequest)

			By("then expecting correct HTTP status code")
			Expect(httpResponseRecorder.Code).To(Equal(200))
		})

		It("should produce a handler with a remove route", func() {
			By("faking out the driver")
			driver := &volmanfakes.FakeDriver{}
			driver.RemoveReturns(voldriver.ErrorResponse{})
			handler, err := driverhttp.NewHandler(testLogger, driver)
			Expect(err).NotTo(HaveOccurred())

			httpResponseRecorder := httptest.NewRecorder()
			removeRequest := voldriver.RemoveRequest{Name: "some-volume"}
			removeJSONRequest, err := json.Marshal(removeRequest)
			Expect(err).NotTo(HaveOccurred())

			By("then fake serving the response using the handler")
			route, found := voldriver.Routes.FindRouteByName(voldriver.RemoveRoute)
			Expect(found).To(BeTrue())

			path := fmt.Sprintf("http://0.0.0.0%s", route.Path)
			httpRequest, err := http.NewRequest("POST", path, bytes.NewReader(removeJSONRequest))
			Expect(err).NotTo(HaveOccurred())
			handler.ServeHTTP(httpResponseRecorder, httpRequest)

			By("then expecting correct HTTP status code")
			Expect(httpResponseRecorder.Code).To(Equal(200))
		})
	})
})
