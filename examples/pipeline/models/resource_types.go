package models

type ResourceType struct {
	Name   string
	Type   string
	Source map[string]interface{}
}

type ResourceTypes []ResourceType

//type image struct {
//	repository string
//	tag        string
//	privileged bool
//}
//
//func (i image) Repository() string {
//	return fmt.Sprintf("%s:%s", i.repository, i.tag)
//}

func (resourceTypes ResourceTypes) FindByName(name string) *ResourceType {
	for _, r := range resourceTypes {
		if r.Name == name {
			return &r
		}
	}
	return nil
}

//func (r *resourceTypesManager) Add(
//	name string,
//	_type string,
//	source map[string]interface{},
//) {
//	//assume _type is always `docker-image` at the moment
//	tag := "latest"
//	if t, ok := source["tag"]; ok {
//		tag = t.(string)
//	}
//
//	privileged := false
//	if t, ok := source["privileged"]; ok {
//		privileged = t.(bool)
//	}
//
//	r.images[name] = image{
//		repository: name,
//		tag:        tag,
//		privileged: privileged,
//	}
//}
