package volhttp_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
	"io/ioutil"
	"strings"
	"github.com/onsi/gomega/gexec"
)

var tmpDriversPath string
var fakeDriverPath string

func TestHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Handlers Suite")
}
var _ = SynchronizedBeforeSuite(func() []byte {
	var err error
	

	fakeDriverPath, err = gexec.Build("github.com/cloudfoundry-incubator/volman/fakedriver", "-race")
	Expect(err).NotTo(HaveOccurred())
	return []byte(strings.Join([]string{ fakeDriverPath}, ","))
}, func(pathsByte []byte) {
	path := string(pathsByte)
	fakeDriverPath = strings.Split(path, ",")[0]
})

var _ = BeforeEach(func() {
	var err error
	tmpDriversPath, err = ioutil.TempDir("", "driversPath")
	Expect(err).NotTo(HaveOccurred())

})



var _ = SynchronizedAfterSuite(func() {

}, func() {
	gexec.CleanupBuildArtifacts()
})
