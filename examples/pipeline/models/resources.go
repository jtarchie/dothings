package models

type Resource struct {
	Name   string
	Type   string
	Source map[string]interface{}
}

type Resources []Resource

func (resources Resources) FindByName(name string) *Resource {
	for _, resource := range resources {
		if resource.Name == name {
			return &resource
		}
	}
	return nil
}