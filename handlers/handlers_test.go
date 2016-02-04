package handlers_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/volman/handlers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Generate", func() {

	Context("when generated", func() {
		It("should produce handler with list drivers route", func() {
			handler, _ := handlers.New()
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "http://0.0.0.0/v1/drivers", nil)
			handler.ServeHTTP(w, r)
			Expect(w.Body.String()).Should(Equal("{\"drivers\":null}"))
			Expect(w.Code).Should(Equal(200))
		})
	})
})
