package steps_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"

	"github.com/jtarchie/dothings/examples/pipeline/models"
	"github.com/jtarchie/dothings/examples/pipeline/steps"
	"github.com/jtarchie/dothings/examples/pipeline/steps/managers"
	"github.com/jtarchie/dothings/examples/pipeline/steps/stepsfakes"
	"github.com/jtarchie/dothings/status"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetResource", func() {
	It("returns a unique ID", func() {
		check := steps.NewGetResource(
			&models.Resource{
				Name: "testing",
			},
			&stepsfakes.FakeVersionManager{},
			&stepsfakes.FakeVolumeManager{},
			&stepsfakes.FakeContainerManager{},
			nil,
		)
		Expect(check.ID()).To(ContainSubstring("get resource: testing"))
	})

	When("executing the step", func() {
		var (
			get              *steps.GetResource
			resource         *models.Resource
			versionManager   *stepsfakes.FakeVersionManager
			volumeManager    *stepsfakes.FakeVolumeManager
			containerManager *stepsfakes.FakeContainerManager
		)

		BeforeEach(func() {
			resource = &models.Resource{
				Name: "testing",
			}
			versionManager = &stepsfakes.FakeVersionManager{}
			containerManager = &stepsfakes.FakeContainerManager{}
			volumeManager = &stepsfakes.FakeVolumeManager{}
			get = steps.NewGetResource(
				resource,
				versionManager,
				volumeManager,
				containerManager,
				map[string]interface{}{"a": 1},
			)
		})

		When("getting the latest version", func() {
			It("runs successfully", func() {
				s, err := get.Execute(ioutil.Discard, ioutil.Discard)
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(status.Success))
			})
		})

		When("initializing the container", func() {
			It("starts the correct script", func() {
				resource.Type = "mock"

				s, err := get.Execute(ioutil.Discard, ioutil.Discard)
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(status.Success))

				Expect(containerManager.WorkingDirArgsForCall(0)).To(MatchRegexp("/tmp/build/get-"))
				command, args := containerManager.CommandArgsForCall(0)
				Expect(command).To(Equal("/opt/resource/in"))
				Expect(args).To(ConsistOf(MatchRegexp("/tmp/build/get")))
				image, tag := containerManager.ImageArgsForCall(0)
				Expect(image).To(Equal("concourse/mock-resource"))
				Expect(tag).To(Equal("latest"))
			})

			It("passes stdin with the source, version, and params", func() {
				containerManager.RunStub = func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
					input, err := ioutil.ReadAll(stdin)
					Expect(err).NotTo(HaveOccurred())
					Expect(input).To(MatchJSON(`{
						"source": {"config": 1},
						"version": {"ref":"123"},
                        "params": {"a":1}
					}`))

					return nil
				}

				resource.Source = map[string]interface{}{
					"config": 1,
				}
				versionManager.GetLatestVersionReturns(managers.Version{"ref": "123"})

				s, err := get.Execute(ioutil.Discard, ioutil.Discard)
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(status.Success))
			})
		})

		When("the container fails", func() {
			It("fails on an exit code failure", func() {
				containerManager.RunStub = func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
					return &exec.ExitError{}
				}

				s, err := get.Execute(ioutil.Discard, ioutil.Discard)
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(status.Failed))
			})

			It("errors on another error", func() {
				containerManager.RunStub = func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
					return fmt.Errorf("some error")
				}

				s, err := get.Execute(ioutil.Discard, ioutil.Discard)
				Expect(err).To(HaveOccurred())
				Expect(s).To(Equal(status.Errored))
			})
		})
	})
})
