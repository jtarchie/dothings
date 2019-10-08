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

var _ = Describe("CheckResource", func() {
	It("returns a unique ID", func() {
		check := steps.NewCheckResource(
			&models.Resource{
				Name: "testing",
			},
			&stepsfakes.FakeVersionManager{},
			&stepsfakes.FakeContainerManager{},
		)
		Expect(check.ID()).To(ContainSubstring("check resource: testing"))
	})

	When("executing the step", func() {
		var (
			check            *steps.CheckResource
			resource         *models.Resource
			versionManager   *stepsfakes.FakeVersionManager
			containerManager *stepsfakes.FakeContainerManager
		)
		BeforeEach(func() {
			resource = &models.Resource{
				Name: "testing",
			}
			versionManager = &stepsfakes.FakeVersionManager{}
			containerManager = &stepsfakes.FakeContainerManager{}
			check = steps.NewCheckResource(
				resource,
				versionManager,
				containerManager,
			)
		})

		When("there is a new version", func() {
			It("updates the version manager", func() {
				containerManager.RunStub = func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
					stdout.Write([]byte(`[{"ref":"abc"}]`))
					return nil
				}

				s, err := check.Execute(ioutil.Discard, ioutil.Discard)
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(status.Success))

				Expect(versionManager.SetLatestVersionCallCount()).To(Equal(1))
				r, version := versionManager.SetLatestVersionArgsForCall(0)
				Expect(r).To(Equal(resource))
				Expect(version).To(Equal(managers.Version{"ref": "abc"}))
			})
		})

		When("there is no new version", func() {
			It("does nothing", func() {
				containerManager.RunStub = func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
					stdout.Write([]byte(`[]`))
					return nil
				}

				s, err := check.Execute(ioutil.Discard, ioutil.Discard)
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(status.Success))

				Expect(versionManager.SetLatestVersionCallCount()).To(Equal(0))
			})
		})

		When("there is a current version and source", func() {
			It("passes it as stdin to the resource", func() {
				containerManager.RunStub = func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
					input, err := ioutil.ReadAll(stdin)
					Expect(err).NotTo(HaveOccurred())
					Expect(input).To(MatchJSON(`{
						"source": {"config": 1},
						"version": {"ref":"123"}
					}`))

					stdout.Write([]byte(`[]`))
					return nil
				}

				resource.Source = map[string]interface{}{
					"config": 1,
				}
				versionManager.GetLatestVersionReturns(managers.Version{"ref": "123"})

				s, err := check.Execute(ioutil.Discard, ioutil.Discard)
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(status.Success))
			})
		})

		When("the JSON response is invalid", func() {
			It("errors with a message", func() {
				containerManager.RunStub = func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
					stdout.Write([]byte(`*/invalid JSON/*`))
					return nil
				}

				s, err := check.Execute(ioutil.Discard, ioutil.Discard)
				Expect(err).To(HaveOccurred())
				Expect(s).To(Equal(status.Errored))
			})
		})

		When("initializing the container", func() {
			It("starts the correct script", func() {
				containerManager.RunStub = func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
					stdout.Write([]byte(`[]`))
					return nil
				}

				resource.Type = "mock"

				s, err := check.Execute(ioutil.Discard, ioutil.Discard)
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(status.Success))

				Expect(containerManager.WorkingDirArgsForCall(0)).To(MatchRegexp("/tmp/build/check-"))
				command, args := containerManager.CommandArgsForCall(0)
				Expect(command).To(Equal("/opt/resource/check"))
				Expect(args).To(ConsistOf(MatchRegexp("/tmp/build/check-")))
				image, tag := containerManager.ImageArgsForCall(0)
				Expect(image).To(Equal("concourse/mock-resource"))
				Expect(tag).To(Equal("latest"))
			})
		})

		When("the container fails", func() {
			It("fails on an exit code failure", func() {
				containerManager.RunStub = func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
					return &exec.ExitError{}
				}

				s, err := check.Execute(ioutil.Discard, ioutil.Discard)
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(status.Failed))
			})

			It("errors on another error", func() {
				containerManager.RunStub = func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
					return fmt.Errorf("some error")
				}

				s, err := check.Execute(ioutil.Discard, ioutil.Discard)
				Expect(err).To(HaveOccurred())
				Expect(s).To(Equal(status.Errored))
			})
		})
	})
})
