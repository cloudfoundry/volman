package volman_test

import (
	"fmt"

	. "github.com/cloudfoundry-incubator/volman"
	. "github.com/cloudfoundry-incubator/volman/remote"
	. "github.com/cloudfoundry-incubator/volman/volmanfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var remoteHttpClient *FakeRemoteHttpClientInterface
var requestReturns *FakeRequestReturnInterface

var _ = Describe("ListDrivers", func() {
	BeforeEach(func() {
		remoteHttpClient = new(FakeRemoteHttpClientInterface)
		requestReturns = new(FakeRequestReturnInterface)
		remoteHttpClient.GetReturns(requestReturns)
		client = &RemoteClient{RemoteHttpClient_: remoteHttpClient}
	})

	Context("When error on Volman server", func() {
		It("should report an error and not list drivers", func() {

			requestReturns.AndReturnJsonInReturns(fmt.Errorf("any"))
			_, err := client.ListDrivers()
			Ω(err).Should(HaveOccurred())
		})
	})

	Context("when volman has no drivers", func() {
		It("should report empty list of drivers", func() {

			requestReturns.AndReturnJsonInStub = func(jsonResponse interface{}) error {
				_, ok := jsonResponse.(*Drivers)
				if !ok {
					return fmt.Errorf("Structure to hold JSON is wrong type")
				}
				return nil
			}

			drivers, err := client.ListDrivers()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(drivers.Drivers)).To(Equal(0))
		})
	})
})
