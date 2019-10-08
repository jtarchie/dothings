package managers_test

import (
	"archive/tar"
	"github.com/jtarchie/dothings/examples/pipeline/steps/managers"
	"github.com/jtarchie/dothings/examples/pipeline/steps/managers/docker"
	"github.com/jtarchie/dothings/examples/pipeline/steps/managers/docker/dockerfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var _ = Describe("DockerManager", func() {
	It("runs a container with all options", func() {
		workingDir, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		runner := managers.NewDockerManager(docker.DefaultExecutor)
		runner.WorkingDir(workingDir)
		runner.Command("bash", "-c", `cat - && env && pwd && echo "hello stderr" 1>&2`)
		runner.Image("ubuntu", "latest")

		By("ensuring environment variables override")
		runner.EnvVar("A", "3")
		runner.EnvVar("B", "2")
		runner.EnvVar("A", "1")

		stdout, stderr := gbytes.NewBuffer(), gbytes.NewBuffer()
		err = runner.Run(
			strings.NewReader("things from planet called stdin"),
			io.MultiWriter(stdout, GinkgoWriter),
			io.MultiWriter(stderr, GinkgoWriter),
		)

		By("having values written to stdout")
		Eventually(stdout).Should(gbytes.Say("things from planet called stdin"))
		Eventually(stdout).Should(gbytes.Say("A=1"))
		Eventually(stdout).Should(gbytes.Say("B=2"))
		Eventually(stdout).Should(gbytes.Say(workingDir))

		By("having values written to stderr")
		Eventually(stderr).Should(gbytes.Say("hello stderr"))
	})

	When("an OCI image tarball is provided", func() {
		It("imports that image into the registry", func() {
			ociDir, err := ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			executor := &dockerfakes.FakeCommandExecutor{}
			index := 0
			executor.RunStub = func(stdin io.Reader, stdout io.Writer, stderr io.Writer, command string, args ...string) error {
				outputs := map[int]string{
					0: "image.tar",
					2: `[{"RepoTags":["some-image-name"]}]`,
				}

				if _, ok := outputs[index]; ok {
					_, err := stdout.Write([]byte(outputs[index]))
					Expect(err).NotTo(HaveOccurred())
				}

				index++
				return nil
			}

			workingDir, err := ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			runner := managers.NewDockerManager(executor)
			runner.WorkingDir(workingDir)
			runner.Command("bash", "-c", `pwd`)
			runner.ImageFromOCI(ociDir)

			err = runner.Run(
				nil,
				GinkgoWriter,
				GinkgoWriter,
			)
			Expect(err).NotTo(HaveOccurred())

			args := ""
			for _, invocation := range executor.Invocations()["Run"] {
				args += strings.Join(invocation[4].([]string), " ") + "\n"
			}
			Expect(args).To(ContainSubstring("docker load --input image.tar"))
		})
	})

	When("privileged is set", func() {
		It("starts the container in privileged mode", func() {
			executor := &dockerfakes.FakeCommandExecutor{}
			runner := managers.NewDockerManager(executor)
			runner.WorkingDir("/tmp")
			runner.Image("ubuntu", "")
			runner.Command("bash")
			runner.Privileged(true)

			err := runner.Run(
				nil,
				GinkgoWriter,
				GinkgoWriter,
			)
			Expect(err).NotTo(HaveOccurred())
			_, _, _, command, args := executor.RunArgsForCall(0)
			Expect(command).To(Equal("docker"))
			Expect(args).To(ContainElement("--privileged"))
		})
	})

	When("user is set", func() {
		It("starts the container with that user", func() {
			executor := &dockerfakes.FakeCommandExecutor{}
			runner := managers.NewDockerManager(executor)
			runner.WorkingDir("/tmp")
			runner.Image("ubuntu", "")
			runner.Command("bash")
			runner.User("some-user")

			err := runner.Run(
				nil,
				GinkgoWriter,
				GinkgoWriter,
			)
			Expect(err).NotTo(HaveOccurred())
			_, _, _, command, args := executor.RunArgsForCall(0)
			Expect(command).To(Equal("docker"))
			Expect(args).To(ContainElement(`--user=some-user`))
		})
	})

	When("values are not provided", func() {
		It("errors on missing working dir", func() {
			runner := managers.NewDockerManager(noopExecutor)
			runner.Command("bash")
			runner.Image("ubuntu", "")

			err := runner.Run(
				nil,
				GinkgoWriter,
				GinkgoWriter,
			)
			Expect(err).To(HaveOccurred())
		})

		It("errors on missing command", func() {
			runner := managers.NewDockerManager(noopExecutor)
			runner.WorkingDir("/tmp")
			runner.Image("ubuntu", "")

			err := runner.Run(
				nil,
				GinkgoWriter,
				GinkgoWriter,
			)
			Expect(err).To(HaveOccurred())
		})

		It("errors on missing image", func() {
			runner := managers.NewDockerManager(noopExecutor)
			runner.WorkingDir("/tmp")
			runner.Command("bash")

			err := runner.Run(
				nil,
				GinkgoWriter,
				GinkgoWriter,
			)
			Expect(err).To(HaveOccurred())
		})
	})
})

var noopExecutor = docker.Executor(func(
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	command string,
	args ...string,
) error {
	return nil
})

func writeTarball(directory string, files map[string]string) string {
	tarball, err := ioutil.TempFile(directory, "*.tar")
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

	return tarball.Name()
}

func writeFiles(directory string, files map[string]string) {
	for file, contents := range files {
		err := ioutil.WriteFile(filepath.Join(directory, file), []byte(contents), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	}
}
