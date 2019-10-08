package docker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type importOCI struct {
	volumeName      string
	commandExecutor CommandExecutor
}

func NewImportOCI(
	directory string,
	commandExecutor CommandExecutor,
) *importOCI {
	return &importOCI{
		volumeName:      directory,
		commandExecutor: commandExecutor,
	}
}

type dockerEnv struct {
	ImageName string
	Env       []string
	User      string
}

func (i *importOCI) Execute(stdout io.Writer, stderr io.Writer) (*dockerEnv, error) {
	matches := i.glob("*.tar", stdout, stderr)
	if len(matches) > 0 {
		return i.importTarball(matches[0], stdout, stderr)
	}
	matches = i.glob("rootfs/", stdout, stderr)
	if len(matches) > 0 {
		return i.importRootFSDirectory(stdout, stderr)
	}

	return nil, fmt.Errorf("no image tarball could be found in %s", i.volumeName)
}

func (i *importOCI) run(
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	command string,
	args ...string,
) error {
	dndArgs := []string{
		"run", "-i", "--rm",
		"-v", "/var/run/docker.sock:/var/run/docker.sock",
		"-v", fmt.Sprintf("%s:/%s", i.volumeName, i.volumeName),
		"-w", fmt.Sprintf("/%s", i.volumeName),
		"--privileged",
		"docker", command,
	}
	dndArgs = append(dndArgs, args...)
	return i.commandExecutor.Run(
		stdin, stdout, stderr,
		"docker", dndArgs...,
	)
}

func (i *importOCI) importTarball(tarballFilename string, stdout io.Writer, stderr io.Writer) (*dockerEnv, error) {
	err := i.run(
		nil,
		stdout,
		stderr,
		"docker",
		"load", "--input",
		tarballFilename,
	)
	if err != nil {
		return nil, fmt.Errorf("could not load image file %s: %s", tarballFilename, err)
	}
	imageName, err := i.getImageNameFromManifest(tarballFilename, stderr)
	if err != nil {
		return nil, err
	}
	return &dockerEnv{
		ImageName: imageName,
	}, nil
}

func (i *importOCI) importRootFSDirectory(stdout io.Writer, stderr io.Writer) (*dockerEnv, error) {
	contents, err := i.readFile("metadata.json", stderr)
	if err != nil {
		return nil, fmt.Errorf("could not load metadata.json: %s", err)
	}
	var metadata struct {
		User string
		Env  []string
	}
	err = json.Unmarshal(contents, &metadata)
	if err != nil {
		return nil, fmt.Errorf("could not json unmarshal metadata.json: %s", err)
	}

	imageNameBytes := &bytes.Buffer{}
	err = i.run(
		nil,
		io.MultiWriter(stdout, imageNameBytes),
		stderr,
		"sh",
		"-c", "set -e; tar cf - -C rootfs/ . | docker import -",
	)
	if err != nil {
		return nil, fmt.Errorf("could not rootfs/ volume : %s", err)
	}

	return &dockerEnv{
		ImageName: strings.TrimSpace(imageNameBytes.String()),
		Env:       metadata.Env,
		User:      metadata.User,
	}, nil
}

func (i *importOCI) glob(pattern string, stdout io.Writer, stderr io.Writer) []string {
	listing := &bytes.Buffer{}
	err := i.run(
		nil,
		io.MultiWriter(stdout, listing),
		stderr,
		"sh", "-c", fmt.Sprintf("ls -1a %s", pattern),
	)
	if err != nil {
		return nil
	}

	matches := strings.Split(listing.String(), "\n")
	if len(matches) == 1 && matches[0] == "" {
		return nil
	}

	return matches
}

func (i *importOCI) readFile(path string, stderr io.Writer) ([]byte, error) {
	contents := &bytes.Buffer{}
	err := i.run(
		nil,
		io.MultiWriter(contents, stderr),
		stderr,
		"cat", path,
	)
	if err != nil {
		return nil, err
	}
	return contents.Bytes(), nil
}

func (i *importOCI) getImageNameFromManifest(tarballFilename string, stderr io.Writer) (string, error) {
	contents := &bytes.Buffer{}
	err := i.run(
		nil, contents, stderr,
		"tar", "-xOf", tarballFilename, "manifest.json",
	)
	if err != nil {
		return "", fmt.Errorf("could not read contents of manifest.json: %s", err)
	}

	var payload []struct {
		RepoTags []string
	}
	err = json.Unmarshal(contents.Bytes(), &payload)
	if err != nil {
		return "", fmt.Errorf("could not unmarhsal JSON of manifest.json: %s", err)
	}
	if len(payload) > 0 {
		return payload[0].RepoTags[0], nil
	}
	return "", fmt.Errorf("could not find repo tag in manifest.json")
}
