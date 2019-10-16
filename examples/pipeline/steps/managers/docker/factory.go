package docker

import "github.com/jtarchie/dothings/examples/pipeline/steps"

type factory struct {
	resourceVolumeManager *resourceVolumeManager
}

func NewFactory() *factory {
	return &factory{
		resourceVolumeManager: NewResourceVolumeManager(DefaultExecutor),
	}
}

func (f *factory) VolumeManager() steps.VolumeManager {
	return f.resourceVolumeManager
}

func (f *factory) NewContainerManager() steps.ContainerManager {
	return nil
}
