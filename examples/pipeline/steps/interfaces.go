package steps

import (
	"io"

	"github.com/jtarchie/dothings/examples/pipeline/models"
	"github.com/jtarchie/dothings/examples/pipeline/steps/managers"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . versionManager
type versionManager interface {
	GetLatestVersion(*models.Resource) managers.Version
	SetLatestVersion(*models.Resource, managers.Version)
}

//counterfeiter:generate . VolumeManager
type VolumeManager interface {
	Get(string, bool) string
	All() map[string]string
}

//counterfeiter:generate . ContainerManager
type ContainerManager interface {
	Volume(local string, mountAs string)
	WorkingDir(dir string)
	Command(command string, args ...string)
	Image(name string, tag string)
	ImageFromOCI(directory string)
	EnvVar(name string, value string)
	Privileged(bool)
	User(string)
	Run(
		stdin io.Reader,
		stdout io.Writer,
		stderr io.Writer,
	) error
}


//counterfeiter:generate . factory
type factory interface {
	VolumeManager() VolumeManager
	NewContainerManager() ContainerManager
}