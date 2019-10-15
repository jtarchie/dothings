package models

type ResourceType struct {
	Name   string
	Type   string
	Source map[string]interface{}
}

type ResourceTypes []ResourceType

func (resourceTypes ResourceTypes) FindByName(name string) *ResourceType {
	for i := len(resourceTypes) - 1; i >= 0; i-- {
		r := resourceTypes[i]

		if r.Name == name {
			return &r
		}
	}
	return nil
}

func (resourceTypes *ResourceTypes) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r := []ResourceType{}

	err := unmarshal(&r)
	if err != nil {
		return err
	}

	*resourceTypes = append(*resourceTypes, r...)

	return nil
}