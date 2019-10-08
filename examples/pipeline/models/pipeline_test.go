package models_test

import (
	. "github.com/jtarchie/dothings/examples/pipeline/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

var _ = Describe("Parsing a pipeline from YAML", func() {
	It("handles parsing tasks, get, and put", func() {
		var pipeline Pipeline
		err := yaml.UnmarshalStrict(doc, &pipeline)
		Expect(err).NotTo(HaveOccurred())

		Expect(len(pipeline.Resources)).To(BeNumerically(">=", 1))
		Expect(len(pipeline.Jobs)).To(BeNumerically(">=", 1))
		//Expect(len(pipeline.ResourceTypes)).To(BeNumerically(">=", 1))
	})
})
