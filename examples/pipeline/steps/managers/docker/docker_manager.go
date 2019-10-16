package docker

import (
	"fmt"
	"io"
	"sort"
)

type dockerManager struct {
	workingDir      string
	volumes         map[string]string
	command         []string
	image           string
	imageFromOCIDir string
	env             map[string]string
	commandExecutor CommandExecutor
	privileged      bool
	user            string
}

func (d *dockerManager) Volume(local string, mountAs string) {
	d.volumes[local] = mountAs
}

func (d *dockerManager) WorkingDir(dir string) {
	d.workingDir = dir
}

func (d *dockerManager) Command(command string, args ...string) {
	d.command = []string{command}
	d.command = append(d.command, args...)
}

func (d *dockerManager) Image(name string, tag string) {
	if tag == "" {
		d.image = name
	} else {
		d.image = fmt.Sprintf("%s:%s", name, tag)
	}
}

func (d *dockerManager) ImageFromOCI(directory string) {
	d.imageFromOCIDir = directory
}

func (d *dockerManager) EnvVar(name string, value string) {
	d.env[name] = value
}

var (
	ErrWorkingDirectory = fmt.Errorf("working directory is required")
	ErrCommandRequired  = fmt.Errorf("command is required")
	ErrImageRequired    = fmt.Errorf("image is required")
)

func (d *dockerManager) Run(
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
) error {
	if d.workingDir == "" {
		return ErrWorkingDirectory
	}

	if len(d.command) == 0 {
		return ErrCommandRequired
	}

	args := []string{
		"run", "-i", "--rm",
		"-w", d.workingDir,
		"--entrypoint", "",
	}

	for localPath, mountPath := range d.volumes {
		args = append(args,
			fmt.Sprintf("-v=%s:%s", localPath, mountPath),
		)
	}

	var imageName string

	if len(d.image) > 0 {
		imageName = d.image
	} else if len(d.imageFromOCIDir) > 0 {
		importer := NewImportOCI(
			d.imageFromOCIDir,
			d.commandExecutor,
		)
		env, err := importer.Execute(stdout, stderr)
		if err != nil {
			return err
		}
		for _, envVar := range env.Env {
			args = append(args, fmt.Sprintf("-e=%s", envVar))
		}
		if env.User != "" && d.user == "" {
			args = append(args, fmt.Sprintf("--user=%s", env.User))
		}
		imageName = env.ImageName
	} else {
		return ErrImageRequired
	}

	envNames := []string{}
	for name, _ := range d.env {
		envNames = append(envNames, name)
	}
	sort.Strings(envNames)

	for _, name := range envNames {
		args = append(args,
			fmt.Sprintf("-e=%s=%s", name, d.env[name]),
		)
	}

	if d.privileged {
		args = append(args, "--privileged")
	}

	if d.user != "" {
		args = append(args, fmt.Sprintf("--user=%s", d.user))
	}

	args = append(args, imageName)
	args = append(args, d.command...)

	return d.commandExecutor.Run(
		stdin,
		stdout,
		stderr,
		"docker",
		args...,
	)
}

func (d *dockerManager) Privileged(b bool) {
	d.privileged = b
}

func (d *dockerManager) User(s string) {
	d.user = s
}

func NewDockerManager(runner CommandExecutor) *dockerManager {
	return &dockerManager{
		volumes:         map[string]string{},
		env:             map[string]string{},
		commandExecutor: runner,
	}
}
