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
	Context("after starting", func() {
		BeforeEach(func() {
			volmanProcess = ginkgomon.Invoke(runner)
			time.Sleep(time.Millisecond * 1000)
		})

		It("it should not exit", func() {
			Consistently(runner).ShouldNot(Exit())
		})
		It("it should get a 404 for root", func() {
			time.Sleep(time.Millisecond * 9000)
			req, _ := http.NewRequest("GET", fmt.Sprintf("http://0.0.0.0:%d/", volmanServerPort), nil)
			response, _ := (&http.Client{}).Do(req)
			defer response.Body.Close()
			body, err := ioutil.ReadAll(response.Body)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(body).Should(ContainSubstring("404"))
		})
	})

})
