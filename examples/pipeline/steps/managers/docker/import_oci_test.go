package docker_test

import (
	"errors"
	"fmt"
	"github.com/jtarchie/dothings/examples/pipeline/steps/managers/docker"
	"github.com/jtarchie/dothings/examples/pipeline/steps/managers/docker/dockerfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
	"io/ioutil"
	"strings"
)

var _ = Describe("ImportOci", func() {
	When("a rootfs directory exists", func() {
		var (
			directory string
		)

		BeforeEach(func() {
			var err error
			directory, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())
		})

		It("loads it into docker", func() {
			executor := &dockerfakes.FakeCommandExecutor{}
			setupDockerExecution(executor, map[string]interface{}{
				"ls -1a rootfs": "rootfs",
				"docker import": "sha256:asdf",
				"metadata.json": `{"user":"root","env":["PATH=/bin","HOME=/root"]}`,
			})

			importer := docker.NewImportOCI(directory, executor)

			env, err := importer.Execute(ioutil.Discard, ioutil.Discard)
			Expect(err).NotTo(HaveOccurred())
			Expect(env.ImageName).To(Equal("sha256:asdf"))
			Expect(env.Env).To(Equal([]string{"PATH=/bin", "HOME=/root"}))
			Expect(env.User).To(Equal("root"))

			invoked := invocationsAsString(executor)
			Expect(invoked).To(ContainSubstring(`sh -c set -e; tar cf - -C rootfs/ . | docker import -`))
		})
	})

	When("a image tarball exists", func() {
		var (
			directory string
		)

		BeforeEach(func() {
			var err error
			directory, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())
		})

		It("loads it into docker", func() {
			executor := &dockerfakes.FakeCommandExecutor{}
			setupDockerExecution(executor,
				map[string]interface{}{
					"*.tar":         "image.tar",
					"manifest.json": `[{"RepoTags":["some-image-name"]}]`,
				},
			)

			importer := docker.NewImportOCI(directory, executor)
			env, err := importer.Execute(ioutil.Discard, ioutil.Discard)
			Expect(err).NotTo(HaveOccurred())
			Expect(env.ImageName).To(Equal("some-image-name"))

			invokes := invocationsAsString(executor)

			Expect(invokes).To(ContainSubstring("docker load --input image.tar"))
		})

		When("the image cannot be loaded into docker", func() {
			It("errors with a helpful message", func() {
				executor := &dockerfakes.FakeCommandExecutor{}
				setupDockerExecution(executor,
					map[string]interface{}{
						"*.tar":         "image.tar",
						"manifest.json": `[{"RepoTags":["some-image-name"]}]`,
						"docker load":   errors.New("human errored"),
					},
				)

				importer := docker.NewImportOCI(directory, executor)

				env, err := importer.Execute(ioutil.Discard, ioutil.Discard)
				Expect(env).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("could not load image file image.tar: human errored"))
			})
		})

		When("manifest.json contains multiple repo tags", func() {
			It("returns the first one", func() {
				executor := &dockerfakes.FakeCommandExecutor{}
				setupDockerExecution(executor,
					map[string]interface{}{
						"*.tar":         "image.tar",
						"manifest.json": `[{"RepoTags":["some-image-name","another-name"]},{"RepoTags":["what"]}]`,
					},
				)
				importer := docker.NewImportOCI(directory, executor)

				env, err := importer.Execute(ioutil.Discard, ioutil.Discard)
				Expect(err).NotTo(HaveOccurred())
				Expect(env.ImageName).To(Equal("some-image-name"))
			})
		})

		When("the manifest.json contains no repo tags", func() {
			It("returns a helpful error message", func() {
				executor := &dockerfakes.FakeCommandExecutor{}
				setupDockerExecution(executor,
					map[string]interface{}{
						"*.tar":         "image.tar",
						"manifest.json": `[]`,
					},
				)
				importer := docker.NewImportOCI(directory, executor)

				env, err := importer.Execute(ioutil.Discard, ioutil.Discard)
				Expect(env).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("could not find repo tag in manifest.json"))
			})
		})

		When("the manifest.json is invalid JSON", func() {
			It("returns a helpful error message", func() {
				executor := &dockerfakes.FakeCommandExecutor{}
				setupDockerExecution(executor,
					map[string]interface{}{
						"*.tar":         "image.tar",
						"manifest.json": `invalid JSON`,
					},
				)
				importer := docker.NewImportOCI(directory, executor)

				env, err := importer.Execute(ioutil.Discard, ioutil.Discard)
				Expect(env).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("could not unmarhsal JSON of manifest.json: invalid character 'i' looking for beginning of value"))
			})
		})

		When("the manifest.json does not exist", func() {
			It("returns a helpful error message", func() {
				executor := &dockerfakes.FakeCommandExecutor{}
				setupDockerExecution(executor,
					map[string]interface{}{
						"*.tar":         "image.tar",
						"manifest.json": errors.New("cannot load file"),
					},
				)
				importer := docker.NewImportOCI(directory, executor)

				env, err := importer.Execute(ioutil.Discard, ioutil.Discard)
				Expect(env).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("could not read contents of manifest.json: cannot load file"))
			})
		})
	})

	When("a tarball does not exist", func() {
		It("gives a helpful error", func() {
			executor := &dockerfakes.FakeCommandExecutor{}
			importer := docker.NewImportOCI("some-random-directory", executor)

			env, err := importer.Execute(ioutil.Discard, ioutil.Discard)
			Expect(env).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("no image tarball could be found in some-random-directory"))
		})
	})
})

func invocationsAsString(executor *dockerfakes.FakeCommandExecutor) string {
	invocations, ok := executor.Invocations()["Run"]
	Expect(ok).To(BeTrue(), "could not get invocations for Run")
	var invokes string
	for _, invocation := range invocations {
		args, ok := invocation[4].([]string)
		Expect(ok).To(BeTrue(), "could not get arguments from invocation of runner")

		invokes += fmt.Sprintf("%s\n", strings.Join(args, " "))
	}
	return invokes
}

func setupDockerExecution(executor *dockerfakes.FakeCommandExecutor, outputs map[string]interface{}) {
	executor.RunStub = func(reader io.Reader, stdout io.Writer, writer io.Writer, command string, args ...string) error {
		invoked := fmt.Sprintf("%s %s", command, strings.Join(args, " "))

		for terms, output := range outputs {
			if strings.Contains(invoked, terms) {
				switch v := output.(type) {
				case string:
					_, err := stdout.Write([]byte(v))
					Expect(err).NotTo(HaveOccurred())
					return nil
				case error:
					return v
				}
			}
		}

		return nil
	}
}
