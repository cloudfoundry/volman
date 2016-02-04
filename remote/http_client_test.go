package remote_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	. "github.com/cloudfoundry-incubator/volman/remote"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type errCloser struct{ io.Reader }

func (errCloser) Close() error                     { return nil }
func (errCloser) Read(p []byte) (n int, err error) { return 0, fmt.Errorf("any") }

type stringCloser struct{ io.Reader }

func (stringCloser) Close() error { return nil }

var _ = Describe("RequestReturn", func() {
	Context("Wrapped errror", func() {
		It("should report an error", func() {
			req := RequestReturn{nil, fmt.Errorf("any")}
			err := req.AndReturnJsonIn(nil)
			Expect(err).Should(HaveOccurred())
		})
	})
	Context("Wrapped response without body", func() {
		It("should report an error", func() {
			req := RequestReturn{&http.Response{Body: errCloser{bytes.NewBufferString("")}}, nil}
			err := req.AndReturnJsonIn(nil)
			Expect(err).Should(HaveOccurred())
		})
	})
	Context("Wrapped response with body that is not correct JSON", func() {
		It("should report an error", func() {
			req := RequestReturn{&http.Response{Body: stringCloser{bytes.NewBufferString("")}}, nil}
			err := req.AndReturnJsonIn(nil)
			Expect(err).Should(HaveOccurred())
		})
	})
	Context("Wrapped response with body that is correct JSON", func() {
		It("should fill in structure", func() {
			req := RequestReturn{&http.Response{Body: stringCloser{bytes.NewBufferString("{\"string\":\"value\"}")}}, nil}
			structure := struct {
				String string `json:"string"`
			}{}
			err := req.AndReturnJsonIn(&structure)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(structure.String).Should(Equal("value"))
		})
	})
})

type MockClient struct{}

var MyClient func() (resp *http.Response, err error)

func (c *MockClient) Do(req *http.Request) (resp *http.Response, err error) {
	Expect(req.URL.Path).Should(Equal("/"))
	resp, err = MyClient()
	return
}

var _ = Describe("Volman Http Client", func() {
	Context("Getting request", func() {
		It("should get a response", func() {
			MyClient = func() (resp *http.Response, err error) { return nil, nil }
			v := RemoteHttpClient{HttpClient: &MockClient{}}
			response := v.Get("/")
			Expect(response).ShouldNot(BeNil())
		})
	})
})
