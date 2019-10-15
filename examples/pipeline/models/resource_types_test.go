package models_test

import (
	"github.com/jtarchie/dothings/examples/pipeline/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

var _ = Describe("Resource Types", func() {
	It("can find resource type by name", func() {
		resource_types := models.ResourceTypes{
			{Name: "A"},
			{Name: "B"},
		}
		Expect(resource_types.FindByName("A").Name).To(Equal("A"))
		Expect(resource_types.FindByName("B").Name).To(Equal("B"))
		Expect(resource_types.FindByName("C")).To(BeNil())
	})

	When("initialize with a resource type", func() {
		It("can be overridden by the pipeline", func() {
			pipeline := models.NewPipeline(
				models.ResourceTypes{
					{Name: "A"},
					{Name: "B"},
				},
			)
			config := `
resource_types:
- name: A
  type: some-type
`
			err := yaml.UnmarshalStrict([]byte(config), &pipeline)
			Expect(err).NotTo(HaveOccurred())

			resourceType := pipeline.ResourceTypes.FindByName("A")
			Expect(resourceType.Name).To(Equal("A"))
			Expect(resourceType.Type).To(Equal("some-type"))

			Expect(pipeline.ResourceTypes.FindByName("B").Name).To(Equal("B"))
			Expect(pipeline.ResourceTypes.FindByName("C")).To(BeNil())
		})
	})
})
