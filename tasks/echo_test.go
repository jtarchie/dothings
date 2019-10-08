package tasks_test

import (
	"github.com/jtarchie/dothings/status"
	"github.com/jtarchie/dothings/tasks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Echo", func() {
	It("returns the ID of the of the original name", func() {
		task := tasks.NewEcho("Hello, World", status.Success)
		Expect(task.ID()).To(Equal("Hello, World"))
	})

	It("writes an execution to stdout and stderr", func() {
		task := tasks.NewEcho("1", status.Success)
		stdout, stderr := gbytes.NewBuffer(), gbytes.NewBuffer()
		Expect(task.Execute(stdout, stderr)).To(Equal(status.Success))
		Expect(stdout).To(gbytes.Say("out: executing 1"))
		Expect(stderr).To(gbytes.Say("err: executing 1"))
	})

	It("returns the specific state", func() {
		task := tasks.NewEcho("1", status.Failed)
		stdout, stderr := gbytes.NewBuffer(), gbytes.NewBuffer()
		Expect(task.Execute(stdout, stderr)).To(Equal(status.Failed))
	})
})
