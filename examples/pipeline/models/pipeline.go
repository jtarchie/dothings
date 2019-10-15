package models

type Pipeline struct {
	Resources     Resources     `yaml:"resources,omitempty"`
	ResourceTypes ResourceTypes `yaml:"resource_types,omitempty"`
	Jobs          Jobs          `yaml:"jobs,omitempty"`
}

func NewPipeline(resourceTypes ResourceTypes) *Pipeline {
	return &Pipeline{
		ResourceTypes: resourceTypes,
	}
}
