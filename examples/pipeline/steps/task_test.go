package steps_test

import (
	"fmt"
	"github.com/onsi/gomega/gbytes"
	"io"
	"io/ioutil"
	"os/exec"

	"github.com/jtarchie/dothings/examples/pipeline/models"
	"github.com/jtarchie/dothings/examples/pipeline/steps"
	"github.com/jtarchie/dothings/examples/pipeline/steps/stepsfakes"
	"github.com/jtarchie/dothings/status"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

const validTask = `
jobs:
- name: test
  plan:
  - task: testing
    config:
      platform: linux
      image_resource:
        type: docker-image
        source:
          repository: ubuntu
      inputs:
      - name: resource-a
      - name: output-1
        path: named-input
      outputs:
      - name: output-2
        path: named-output
      params:
        B: 2
        A: 1
      run:
        path: echo
        args: ["hello world"]
`

var _ = Describe("Task", func() {
	It("returns a unique ID", func() {
		check := steps.NewTask(
			newTask(validTask),
			&stepsfakes.FakeVolumeManager{},
			&stepsfakes.FakeContainerManager{},
		)
		Expect(check.ID()).To(ContainSubstring("task: testing"))
	})

	When("executing valid task", func() {
		var (
			task             *steps.Task
			volumeManager    *stepsfakes.FakeVolumeManager
			containerManager *stepsfakes.FakeContainerManager
		)

		BeforeEach(func() {
			volumeManager = &stepsfakes.FakeVolumeManager{}
			containerManager = &stepsfakes.FakeContainerManager{}
			task = steps.NewTask(
				newTask(validTask),
				volumeManager,
				containerManager,
			)
		})

		It("is successful on completion", func() {
			s, err := task.Execute(ioutil.Discard, ioutil.Discard)
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal(status.Success))
		})

		It("set the correct working directory", func() {
			_, _ = task.Execute(ioutil.Discard, ioutil.Discard)
			Expect(containerManager.WorkingDirArgsForCall(0)).To(MatchRegexp(`/tmp/build/\w{6}`))
		})

		It("setups inputs in the working directory", func() {
			_, _ = task.Execute(ioutil.Discard, ioutil.Discard)

			from, to := containerManager.VolumeArgsForCall(0)
			Expect(from).To(Equal(""))
			Expect(to).To(MatchRegexp(`/tmp/build/\w{6}/resource-a`))

			from, to = containerManager.VolumeArgsForCall(1)
			Expect(from).To(Equal(""))
			Expect(to).To(MatchRegexp(`/tmp/build/\w{6}/named-input`))
		})

		It("setups outputs in the working directory", func() {
			_, _ = task.Execute(ioutil.Discard, ioutil.Discard)

			from, to := containerManager.VolumeArgsForCall(2)
			Expect(from).To(Equal(""))
			Expect(to).To(MatchRegexp(`/tmp/build/\w{6}/named-output`))
		})

		It("sets the params as environment variables in sorted order", func() {
			_, _ = task.Execute(ioutil.Discard, ioutil.Discard)

			name, value := containerManager.EnvVarArgsForCall(1)
			Expect(name).To(Equal("B"))
			Expect(value).To(Equal("2"))

			name, value = containerManager.EnvVarArgsForCall(0)
			Expect(name).To(Equal("A"))
			Expect(value).To(Equal("1"))
		})

		It("uses the specified docker image", func() {
			_, _ = task.Execute(ioutil.Discard, ioutil.Discard)

			image, tag := containerManager.ImageArgsForCall(0)
			Expect(image).To(Equal("ubuntu"))
			Expect(tag).To(Equal("latest"))
		})

		It("gives access to stdout and stderr", func() {
			containerManager.RunStub = func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
				stdout.Write([]byte("hello stdout"))
				stderr.Write([]byte("hello stderr"))
				return nil
			}

			stdout, stderr := gbytes.NewBuffer(), gbytes.NewBuffer()
			_, _ = task.Execute(stdout, stderr)

			Eventually(stdout).Should(gbytes.Say("hello stdout"))
			Eventually(stderr).Should(gbytes.Say("hello stderr"))
		})

		It("runs a command", func() {
			containerManager.RunStub = func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
				return nil
			}

			_, _ = task.Execute(ioutil.Discard, ioutil.Discard)
			command, args := containerManager.CommandArgsForCall(0)
			Expect(command).To(Equal("echo"))
			Expect(args).To(Equal([]string{"hello world"}))
		})

		When("the container fails", func() {
			It("fails on an exit code failure", func() {
				containerManager.RunStub = func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
					return &exec.ExitError{}
				}

				s, err := task.Execute(ioutil.Discard, ioutil.Discard)
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(status.Failed))
			})

			It("errors on another error", func() {
				containerManager.RunStub = func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
					return fmt.Errorf("some error")
				}

				s, err := task.Execute(ioutil.Discard, ioutil.Discard)
				Expect(err).To(HaveOccurred())
				Expect(s).To(Equal(status.Errored))
			})
		})
	})

	When("an image is provided", func() {
		const taskWithImage = `
jobs:
- name: test
  plan:
  - task: testing
    image: resource
    config:
      platform: linux
`

		var (
			task             *steps.Task
			volumeManager    *stepsfakes.FakeVolumeManager
			containerManager *stepsfakes.FakeContainerManager
		)

		BeforeEach(func() {
			volumeManager = &stepsfakes.FakeVolumeManager{}
			containerManager = &stepsfakes.FakeContainerManager{}
			task = steps.NewTask(
				newTask(taskWithImage),
				volumeManager,
				containerManager,
			)
		})

		It("imports that image into the container manager", func() {
			tarballDir, err := ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())
			_, err = ioutil.TempFile(tarballDir, ".tar")
			Expect(err).NotTo(HaveOccurred())

			volumeManager.GetReturnsOnCall(0, tarballDir)

			s, err := task.Execute(ioutil.Discard, ioutil.Discard)
			Expect(err).NotTo(HaveOccurred())
			Expect(s).To(Equal(status.Success))

			Expect(containerManager.ImageFromOCIArgsForCall(0)).To(Equal(tarballDir))
		})
	})

	XWhen("a docker image resource is provided", func() {
		It("uses the repository for the container", func() {

		})
	})

	XWhen("a non-docker image resource is provided", func() {
		It("errors with unsupported", func() {

		})
	})
})

func newTask(source string) models.Step {
	var pipeline models.Pipeline
	err := yaml.UnmarshalStrict([]byte(source), &pipeline)
	Expect(err).NotTo(HaveOccurred())
	return pipeline.Jobs.FindByName("test").Steps[0]
}
