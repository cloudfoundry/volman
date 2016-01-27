package volman_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/cloudfoundry-incubator/volman/handlers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Generate", func() {

	Context("when generated", func() {
		It("should produce handler with list drivers route", func() {
			handler, _ := Generate()
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "http://0.0.0.0/v1/drivers", nil)
			handler.ServeHTTP(w, r)
			Ω(w.Body.String()).Should(Equal("{\"drivers\":null}"))
			Ω(w.Code).Should(Equal(200))
		})
	})
})
