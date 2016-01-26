package main_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Volman", func() {
	BeforeEach(func() {
		volmanProcess = ginkgomon.Invoke(runner)
		time.Sleep(time.Millisecond * 1000)
	})

	Context("after starting", func() {
		It("should not exit", func() {
			Consistently(runner).ShouldNot(Exit())
		})
	})
	Context("after starting http server", func() {
		BeforeEach(func() {
			time.Sleep(time.Millisecond * 9000)
		})

		It("should get a 404 for root", func() {
			_, status, err := get("/")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(status).Should(ContainSubstring("404"))
		})
		It("should return empty list for '/v1/drivers' (200 status)", func() {
			body, status, err := get("/v1/drivers")
			fmt.Printf("%s", body)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(status).Should(ContainSubstring("200"))
			Ω(body).Should(ContainSubstring("none"))
		})
	})
})

func get(path string) (body string, status string, err error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://0.0.0.0:%d%s", volmanServerPort, path), nil)
	response, _ := (&http.Client{}).Do(req)
	defer response.Body.Close()
	bodyBytes, err := ioutil.ReadAll(response.Body)
	return string(bodyBytes[:]), response.Status, err
}
