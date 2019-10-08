package planner_test

import (
	. "github.com/jtarchie/dothings/planner"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/jtarchie/dothings/status"
	"github.com/jtarchie/dothings/tasks"
)

var _ = Describe("Tree", func() {
	var plan Step

	BeforeEach(func() {
		var err error

		plan, err = NewSerial(func(plan Planner) error {
			err := plan.Parallel(func(parallel Planner) error {
				err := parallel.Try(func(try Planner) error {
					try.Task(tasks.NewEcho("a", status.Success))
					return nil
				})
				Expect(err).NotTo(HaveOccurred())

				parallel.Task(tasks.NewEcho("a1", status.Success))
				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			err = plan.Success(func(success Planner) error {
				err := success.Serial(func(serial Planner) error {
					serial.Task(tasks.NewEcho("b", status.Success))
					return nil
				})
				Expect(err).NotTo(HaveOccurred())

				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			err = plan.Failure(func(failure Planner) error {
				err := failure.Serial(func(serial Planner) error {
					serial.Task(tasks.NewEcho("c", status.Success))
					return nil
				})
				Expect(err).NotTo(HaveOccurred())

				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			err = plan.Finally(func(finally Planner) error {
				err := finally.Serial(func(serial Planner) error {
					serial.Task(tasks.NewEcho("d", status.Success))
					return nil
				})
				Expect(err).NotTo(HaveOccurred())

				return nil
			})
			Expect(err).NotTo(HaveOccurred())
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("checks for children nodes", func() {
		tree := plan.Tree()
		Expect(tree.Type()).To(Equal(Serial))
		Expect(tree.Children()).To(HaveLen(4))
		Expect(tree.Children()[0].Type()).To(Equal(Parallel))
		children := tree.Children()[0].Children()
		Expect(children[0].Type()).To(Equal(Try))
		Expect(children[0].Children()[0].Type()).To(Equal(Task))
		Expect(children[0].Children()[0].Task().ID()).To(Equal("a"))
		Expect(children[1].Type()).To(Equal(Task))
		Expect(children[1].Task().ID()).To(Equal("a1"))

		Expect(tree.Children()[1].Type()).To(Equal(Success))
		Expect(tree.Children()[2].Type()).To(Equal(Failure))
		Expect(tree.Children()[3].Type()).To(Equal(Finally))
	})
})
