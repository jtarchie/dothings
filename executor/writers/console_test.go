package writers_test

import (
	"github.com/jtarchie/dothings/executor"
	"github.com/jtarchie/dothings/executor/writers"
	dothings "github.com/jtarchie/dothings/planner"
	"github.com/jtarchie/dothings/status"
	"github.com/jtarchie/dothings/tasks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

type buffer struct {
	*gbytes.Buffer
}

func (b buffer) String() string {
	return string(b.Contents())
}

var _ = Describe("Console", func() {
	It("writes a task stdout and stderr to a single buffer", func() {
		stdout, stderr := buffer{gbytes.NewBuffer()}, buffer{gbytes.NewBuffer()}
		console := writers.NewConsole(stdout, stderr)

		taskA := tasks.NewEcho("task 1", status.Success)
		taskB := tasks.NewEcho("task 2", status.Success)

		plan, _ := dothings.NewSerial(func(plan dothings.Planner) error {
			plan.Task(taskA)
			plan.Task(taskB)
			return nil
		})

		executor.NewExecutor(plan, console).Wait()
		Expect(stderr.Buffer).To(gbytes.Say("initializing task 1\n"))
		Expect(stdout.Buffer).To(gbytes.Say("task 1"))
		Expect(stderr.Buffer).To(gbytes.Say("task 1"))
		Expect(stderr.Buffer).To(gbytes.Say("initializing task 2\n"))
		Expect(stdout.Buffer).To(gbytes.Say("task 2"))
		Expect(stderr.Buffer).To(gbytes.Say("task 2"))

		stdoutA, stderrA := console.GetString(taskA)
		Expect(stdoutA).To(ContainSubstring("task 1"))
		Expect(stderrA).To(ContainSubstring("task 1"))

		stdoutA, stderrA = console.GetString(taskA)
		Expect(stdoutA).To(ContainSubstring("task 1"))
		Expect(stderrA).To(ContainSubstring("task 1"))
	})
})
