package docker_test

import (
	"archive/tar"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDocker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Docker Suite")
}

func writeTarball(directory string, files map[string]string) string {
	tarballPath := filepath.Join(directory, "image.tar")
	tarball, err := os.Create(tarballPath)
	Expect(err).NotTo(HaveOccurred())

	tw := tar.NewWriter(tarball)

	for filename, contents := range files {
		hdr := &tar.Header{
			Name: filename,
			Mode: 0600,
			Size: int64(len(contents)),
		}
		err := tw.WriteHeader(hdr)
		Expect(err).NotTo(HaveOccurred())

		_, err = tw.Write([]byte(contents))
		Expect(err).NotTo(HaveOccurred())
	}
	err = tw.Close()
	Expect(err).NotTo(HaveOccurred())
	err = tarball.Close()
	Expect(err).NotTo(HaveOccurred())

	return tarballPath
}

func writeFiles(directory string, files map[string]string) {
	for file, contents := range files {
		err := ioutil.WriteFile(filepath.Join(directory, file), []byte(contents), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	}
}
