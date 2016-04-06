package volhttp_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/volhttp"

	"github.com/cloudfoundry-incubator/volman/volmanfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("Volman Handlers", func() {

	Context("when generating http handlers", func() {

		It("should produce handler with list drivers route", func() {
			testLogger := lagertest.NewTestLogger("HandlersTest")
			client := &volmanfakes.FakeManager{}
			handler, _ := volhttp.NewHandler(testLogger, client)
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "http://0.0.0.0/drivers", nil)
			handler.ServeHTTP(w, r)
			Expect(w.Body.String()).Should(Equal("{\"drivers\":null}"))
			Expect(w.Code).Should(Equal(200))
		})

		It("should produce handler with mount route", func() {
			testLogger := lagertest.NewTestLogger("HandlersTest")
			client := &volmanfakes.FakeManager{}
			client.MountReturns(volman.MountResponse{"dummy_path"}, nil)
			handler, _ := volhttp.NewHandler(testLogger, client)
			w := httptest.NewRecorder()
			MountRequest := volman.MountRequest{}

			mountJSONRequest, _ := json.Marshal(MountRequest)
			r, _ := http.NewRequest("POST", "http://0.0.0.0/drivers/mount", bytes.NewReader(mountJSONRequest))
			handler.ServeHTTP(w, r)
			mountResponse := volman.MountResponse{}
			body, err := ioutil.ReadAll(w.Body)
			err = json.Unmarshal(body, &mountResponse)
			Expect(err).ToNot(HaveOccurred())
			Expect(mountResponse.Path).Should(Equal("dummy_path"))
		})

		It("should produce handler with unmount route", func() {
			testLogger := lagertest.NewTestLogger("HandlersTest")
			client := &volmanfakes.FakeManager{}
			client.UnmountReturns(nil)
			handler, _ := volhttp.NewHandler(testLogger, client)
			w := httptest.NewRecorder()
			unmountRequest := volman.UnmountRequest{}
			unmountJSONRequest, _ := json.Marshal(unmountRequest)
			r, _ := http.NewRequest("POST", "http://0.0.0.0/drivers/unmount", bytes.NewReader(unmountJSONRequest))
			handler.ServeHTTP(w, r)
			Expect(w.Code).To(Equal(200))
		})
	})
})
