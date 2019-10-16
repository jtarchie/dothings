package docker

import (
	"crypto/rand"
	"encoding/base32"
	"log"
	"os"
	"sync"
)

type resourceVolumeManager struct {
	sync.Mutex
	volumes         map[string]string
	commandExecutor CommandExecutor
}

func (vm *resourceVolumeManager) Get(name string, force bool) string {
	vm.Lock()
	defer vm.Unlock()

	volume, ok := vm.volumes[name]
	if ok && !force {
		return volume
	}

	volume = generateVolumeGUID()
	err := vm.commandExecutor.Run(
		nil,
		os.Stderr,
		os.Stderr,
		"docker",
		"volume", "create",
		"--driver", "local",
		volume,
	)
	if err != nil {
		log.Fatal(err)
	}

	vm.volumes[name] = volume
	return volume
}

func (vm *resourceVolumeManager) All() map[string]string {
	vm.Lock()
	defer vm.Unlock()

	return vm.volumes
}

func NewResourceVolumeManager(
	commandExecutor CommandExecutor,
) *resourceVolumeManager {
	return &resourceVolumeManager{
		volumes:         map[string]string{},
		commandExecutor: commandExecutor,
	}
}

func generateVolumeGUID() string {
	buffer := make([]byte, 16)
	_, _ = rand.Reader.Read(buffer)
	return base32.StdEncoding.EncodeToString(buffer)[0:16]
}
