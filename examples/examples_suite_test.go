package main_test

import (
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestExamples(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Examples Suite")
}

var _ = Describe("Examples", func() {
	AfterEach(func() {
		gexec.CleanupBuildArtifacts()
	})

	Context("the console", func() {
		It("works", func() {
			path, err := gexec.Build("github.com/jtarchie/dothings/examples/console", "-race")
			Expect(err).NotTo(HaveOccurred())

			command := exec.Command(path)
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, "10s").Should(gexec.Exit(0))
		})
	})

	Context("the web", func() {
		It("works", func() {
			path, err := gexec.Build("github.com/jtarchie/dothings/examples/web", "-race")
			Expect(err).NotTo(HaveOccurred())

			command := exec.Command(path, "-duration", "100ms", "-polling-interval", "100ms")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session, "5s").Should(gexec.Exit(0))
		})
	})

	Context("the pipeline", func() {
		It("works", func() {
			_, err := gexec.Build("github.com/jtarchie/dothings/examples/pipeline", "-race")
			Expect(err).NotTo(HaveOccurred())
			// not running yet, just want a compiler check
		})
	})
})
