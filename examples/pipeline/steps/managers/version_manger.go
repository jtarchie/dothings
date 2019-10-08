package managers

import (
	"sync"

	"github.com/jtarchie/dothings/examples/pipeline/models"
)

type Version map[string]string

type resourceVersionManager struct {
	sync.Mutex
	version map[*models.Resource]Version
}

func (vm *resourceVersionManager) GetLatestVersion(resource *models.Resource) Version {
	vm.Lock()
	defer vm.Unlock()

	return vm.version[resource]
}

func (vm *resourceVersionManager) SetLatestVersion(resource *models.Resource, v Version) {
	vm.Lock()
	defer vm.Unlock()

	vm.version[resource] = v
}

func NewResourceVersionManager() *resourceVersionManager {
	return &resourceVersionManager{
		version: map[*models.Resource]Version{},
	}
}
