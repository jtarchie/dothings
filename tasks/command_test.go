package tasks_test

import (
	status2 "github.com/jtarchie/dothings/status"
	"github.com/jtarchie/dothings/tasks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Command", func() {
	It("returns the ID as the localCommand being run", func() {
		task := tasks.NewCommand("echo", "Hello World")
		Expect(task.ID()).To(Equal("command: echo Hello World"))
	})

	When("executing the localCommand", func() {
		It("writes to stdout when written to", func() {
			task := tasks.NewCommand("echo", "Hello World")
			stdout, stderr := gbytes.NewBuffer(), gbytes.NewBuffer()
			status, _ := task.Execute(stdout, stderr)

			Expect(status).To(Equal(status2.Success))
			Expect(stdout).To(gbytes.Say("Hello World"))
			Expect(string(stderr.Contents())).To(Equal(""))
		})
		It("writes to stderr when written to", func() {
			task := tasks.NewCommand("bash", "-c", "echo Hello World 1>&2")
			stdout, stderr := gbytes.NewBuffer(), gbytes.NewBuffer()
			status, _ := task.Execute(stdout, stderr)

			Expect(status).To(Equal(status2.Success))
			Expect(stderr).To(gbytes.Say("Hello World"))
			Expect(string(stdout.Contents())).To(Equal(""))
		})
		It("returns failed when program exit code greater than 0", func() {
			task := tasks.NewCommand("false")
			status, _ := task.Execute(GinkgoWriter, GinkgoWriter)
			Expect(status).To(Equal(status2.Failed))

		})
		It("returns success when program exit code is 0", func() {
			task := tasks.NewCommand("true")
			status, _ := task.Execute(GinkgoWriter, GinkgoWriter)
			Expect(status).To(Equal(status2.Success))
		})
	})
})
