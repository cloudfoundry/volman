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

		It("should produce handler with mount route", func() {

			By("faking out the driver")
			driver := &volmanfakes.FakeDriver{}
			driver.MountReturns(voldriver.MountResponse{"dummy_path"}, nil)
			handler, _ := driverhttp.NewHandler(testLogger, driver)
			httpResponseRecorder := httptest.NewRecorder()
			MountRequest := voldriver.MountRequest{}
			mountJSONRequest, _ := json.Marshal(MountRequest)

			By("then fake serving the response using the handler")
			httpRequest, _ := http.NewRequest("POST", "http://0.0.0.0/mount", bytes.NewReader(mountJSONRequest))
			testLogger.Info(fmt.Sprintf("%#v", httpResponseRecorder.Body))
			handler.ServeHTTP(httpResponseRecorder, httpRequest)

			By("then deserialing the HTTP response")
			mountResponse := voldriver.MountResponse{}
			body, err := ioutil.ReadAll(httpResponseRecorder.Body)
			err = json.Unmarshal(body, &mountResponse)

			By("then expecting correct JSON conversion")
			Expect(err).ToNot(HaveOccurred())
			Expect(mountResponse.Path).Should(Equal("dummy_path"))
		})

		It("should produce handler with unmount route", func() {

			By("faking out the driver")
			driver := &volmanfakes.FakeDriver{}
			driver.UnmountReturns(nil)
			handler, _ := driverhttp.NewHandler(testLogger, driver)
			httpResponseRecorder := httptest.NewRecorder()
			unmountRequest := voldriver.UnmountRequest{}
			unmountJSONRequest, _ := json.Marshal(unmountRequest)

			By("then fake serving the response using the handler")
			httpRequest, _ := http.NewRequest("POST", "http://0.0.0.0/unmount", bytes.NewReader(unmountJSONRequest))
			handler.ServeHTTP(httpResponseRecorder, httpRequest)

			By("then expecting correct HTTP status code")
			Expect(httpResponseRecorder.Code).To(Equal(200))

		})
	})
})
