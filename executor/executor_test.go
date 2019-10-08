package executor_test

import (
	"fmt"
	"strings"
	"time"

	"github.com/jtarchie/dothings/executor/writers"
	"github.com/jtarchie/dothings/status"
	"github.com/jtarchie/dothings/tasks"

	"github.com/jtarchie/dothings/executor"
	"github.com/jtarchie/dothings/planner"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

type gstring struct {
	b *gbytes.Buffer
}

func (g gstring) Write(p []byte) (n int, err error) {
	return g.b.Write(p)
}

func (g gstring) String() string {
	return string(g.b.Contents())
}

func (g gstring) Buffer() *gbytes.Buffer {
	return g.b
}

var _ = Describe("Tasker", func() {
	var (
		console executor.Writer
		stdout  gstring
	)

	BeforeEach(func() {
		stdout = gstring{gbytes.NewBuffer()}
		console = writers.NewConsole(stdout, stdout)
	})

	When("there is a single task", func() {
		It("executes then returns", func() {
			a := task("1")
			plan, _ := planner.NewSerial(func(plan planner.Planner) error {
				plan.Task(a)
				return nil
			})

			Expect(executor.NewExecutor(plan, console).Wait()).To(Equal(status.Success))
			Expect(stdout).To(gbytes.Say("executed 1"))
		})

		It("returns the status of the task", func() {
			a := failingTask{"1"}
			plan, _ := planner.NewSerial(func(plan planner.Planner) error {
				plan.Task(a)
				return nil
			})
			Expect(executor.NewExecutor(plan, console).Wait()).To(Equal(status.Failed))
			Expect(stdout).To(gbytes.Say("executed 1"))
		})
	})

	When("tasks are defined in parallel", func() {
		It("runs them in parallel", func() {
			plan, _ := planner.NewParallel(func(plan planner.Planner) error {
				for i := 0; i < 10; i++ {
					plan.Task(timedTask{task(fmt.Sprintf("%dms", 10+i))})
				}
				return nil
			})
			startTime := time.Now()
			Expect(executor.NewExecutor(plan, console).Wait()).To(Equal(status.Success))
			endTime := time.Now()
			Expect(endTime.Sub(startTime)).To(BeNumerically("<", 145*time.Millisecond))
			for i := 0; i < 10; i++ {
				By(fmt.Sprintf("ensuring that %dms only appears once", 10+i))
				Expect(
					strings.Count(
						stdout.String(),
						fmt.Sprintf("executed %dms", 10+i),
					),
				).To(Equal(1))
			}
		})
	})

	When("a task returns an error state", func() {
		It("marks that as the state", func() {
			a := tasks.NewEcho("errored task", status.Errored)

			plan, _ := planner.NewSerial(func(plan planner.Planner) error {
				plan.Task(a)
				return nil
			})

			Expect(executor.NewExecutor(plan, console).Wait()).To(Equal(status.Errored))
			Expect(stdout).To(gbytes.Say("initializing errored task"))
		})
	})

	When("a task returns an error", func() {
		It("marks that as the state", func() {
			a := erroringTask{"1"}

			plan, _ := planner.NewSerial(func(plan planner.Planner) error {
				plan.Task(a)
				return nil
			})

			Expect(executor.NewExecutor(plan, console).Wait()).To(Equal(status.Errored))
			Expect(stdout).To(gbytes.Say("executed 1"))
		})
	})

	When("a completion of one step starts a subsequent step", func() {
		It("does not wait for other running steps to complete", func() {
			a := newBlockingTask("A")
			b := newBlockingTask("B")
			c := newBlockingTask("C")
			d := newBlockingTask("D")

			plan, err := planner.NewParallel(func(plan planner.Planner) error {
				Expect(plan.Serial(func(plan planner.Planner) error {
					plan.Task(a)
					plan.Task(b)
					return nil
				})).NotTo(HaveOccurred())
				Expect(plan.Serial(func(plan planner.Planner) error {
					plan.Task(c)
					plan.Task(d)
					return nil
				})).NotTo(HaveOccurred())
				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			executor := executor.NewExecutor(plan, console)
			go executor.Wait()

			Eventually(stdout.String).Should(ContainSubstring("A"))
			Consistently(stdout.String).ShouldNot(ContainSubstring("B"))
			Eventually(stdout.String).Should(ContainSubstring("C"))
			Consistently(stdout.String).ShouldNot(ContainSubstring("D"))

			a.wait <- struct{}{}

			Eventually(stdout).Should(gbytes.Say("B"))
			Consistently(stdout).ShouldNot(gbytes.Say("D"))

			c.wait <- struct{}{}

			Eventually(stdout).Should(gbytes.Say("D"))

			b.wait <- struct{}{}
			d.wait <- struct{}{}
		})
	})
})
