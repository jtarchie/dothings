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

var _ = Describe("PutResource", func() {
	It("returns a unique ID", func() {
		check := steps.NewPutResource(
			&models.Resource{
				Name: "testing",
			},
			&stepsfakes.FakeVersionManager{},
			&stepsfakes.FakeVolumeManager{},
			&stepsfakes.FakeContainerManager{},
			nil,
		)
		Expect(check.ID()).To(ContainSubstring("put resource: testing"))
	})

	When("executing the step", func() {
		var (
			put              *steps.PutResource
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
			put = steps.NewPutResource(
				resource,
				versionManager,
				volumeManager,
				containerManager,
				map[string]interface{}{"a": 1},
			)
		})

		When("the put is successful", func() {
			It("sets the latest version", func() {
				containerManager.RunStub = func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
					stdout.Write([]byte(`{"version":{"ref":"abcd"}}`))
					return nil
				}

				s, err := put.Execute(ioutil.Discard, ioutil.Discard)
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(status.Success))

				_, version := versionManager.SetLatestVersionArgsForCall(0)
				Expect(version).To(Equal(managers.Version{"ref": "abcd"}))
			})
		})

		When("the container initializes", func() {
			It("starts the correct script", func() {
				containerManager.RunStub = func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
					stdout.Write([]byte(`{"version":{"ref":"abcd"}}`))
					return nil
				}
				resource.Type = "mock"

				s, err := put.Execute(ioutil.Discard, ioutil.Discard)
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(status.Success))

				Expect(containerManager.WorkingDirArgsForCall(0)).To(MatchRegexp("/tmp/build/put-"))
				command, args := containerManager.CommandArgsForCall(0)
				Expect(command).To(Equal("/opt/resource/out"))
				Expect(args).To(ConsistOf(MatchRegexp("/tmp/build/put-")))
				image, tag := containerManager.ImageArgsForCall(0)
				Expect(image).To(Equal("concourse/mock-resource"))
				Expect(tag).To(Equal("latest"))
			})

			It("passes stdin with the source and params", func() {
				containerManager.RunStub = func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
					input, err := ioutil.ReadAll(stdin)
					Expect(err).NotTo(HaveOccurred())
					Expect(input).To(MatchJSON(`{
						"source": {"config": 1},
                        "params": {"a":1}
					}`))
					stdout.Write([]byte(`{"version":{"ref":"abcd"}}`))
					return nil
				}

				resource.Source = map[string]interface{}{
					"config": 1,
				}
				versionManager.GetLatestVersionReturns(managers.Version{"ref": "123"})

				s, err := put.Execute(ioutil.Discard, ioutil.Discard)
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(status.Success))
			})

			When("volumes are available", func() {
				It("mounts all of them", func() {
					volumeManager.AllReturns(map[string]string{
						"resource-a": "/tmp/resource-a",
						"output-1":   "/tmp/output-1",
					})

					containerManager.RunStub = func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
						stdout.Write([]byte(`{"version":{"ref":"abcd"}}`))
						return nil
					}

					s, err := put.Execute(ioutil.Discard, ioutil.Discard)
					Expect(err).ToNot(HaveOccurred())
					Expect(s).To(Equal(status.Success))

					Expect(containerManager.VolumeCallCount()).To(Equal(2))

					from, to := containerManager.VolumeArgsForCall(0)
					Expect(from).To(Equal("/tmp/output-1"))
					Expect(to).To(MatchRegexp("/tmp/build/put-.*?/output-1"))

					from, to = containerManager.VolumeArgsForCall(1)
					Expect(from).To(Equal("/tmp/resource-a"))
					Expect(to).To(MatchRegexp("/tmp/build/put-.*?/resource-a"))
				})
			})
		})

		When("the container fails", func() {
			It("fails on an exit code failure", func() {
				containerManager.RunStub = func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
					return &exec.ExitError{}
				}

				s, err := put.Execute(ioutil.Discard, ioutil.Discard)
				Expect(err).ToNot(HaveOccurred())
				Expect(s).To(Equal(status.Failed))
			})

			It("errors on another error", func() {
				containerManager.RunStub = func(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
					return fmt.Errorf("some error")
				}

				s, err := put.Execute(ioutil.Discard, ioutil.Discard)
				Expect(err).To(HaveOccurred())
				Expect(s).To(Equal(status.Errored))
			})
		})
	})
})
