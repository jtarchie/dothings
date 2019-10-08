package planner_test

import (
	"github.com/jtarchie/dothings/planner"
	"github.com/jtarchie/dothings/status"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/matchers"
	"github.com/onsi/gomega/types"
)

type matcher struct {
	*matchers.EqualMatcher
}

func EqualTasks(expected interface{}) types.GomegaMatcher {
	return &matcher{
		EqualMatcher: &matchers.EqualMatcher{
			Expected: expected,
		},
	}
}

func (m *matcher) Match(actual interface{}) (success bool, err error) {
	convertedActual := []task{}
	for _, e := range actual.(planner.Tasks) {
		convertedActual = append(convertedActual, e.(task))
	}
	return m.EqualMatcher.Match(convertedActual)
}

type fakeState struct {
	statuses map[string][]status.Type
}

func (f *fakeState) Get(task status.Identifier) []status.Type {
	return f.statuses[task.ID()]
}

func (f *fakeState) Add(task status.Identifier, s status.Type) error {
	f.statuses[task.ID()] = append(f.statuses[task.ID()], s)
	return nil
}

func newStatuses() *fakeState {
	return &fakeState{
		statuses: make(map[string][]status.Type),
	}
}

var _ = Describe("Planner", func() {
	Context("a serial plan with single step", func() {
		It("returns the step for executor when it hasn't started yet", func() {
			plan, err := planner.NewSerial(func(plan planner.Planner) error {
				plan.Task(task("A"))
				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(plan.Next(newStatuses())).To(EqualTasks([]task{"A"}))
			Expect(plan.State(newStatuses())).To(Equal(status.Unstarted))
		})

		It("returns no steps when the step has started", func() {
			plan, err := planner.NewSerial(func(plan planner.Planner) error {
				plan.Task(task("A"))
				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			statuses := []status.Type{status.Success, status.Failed, status.Running, status.Errored}
			for _, status := range statuses {
				state := newStatuses()
				Expect(state.Add(task("A"), status)).ToNot(HaveOccurred())

				Expect(plan.Next(state)).To(EqualTasks([]task{}))
				Expect(plan.State(state)).To(Equal(status))
			}
		})
	})

	Context("a parallel plan with a single step", func() {
		It("returns the step for executor when it hasn't started yet", func() {
			plan, err := planner.NewParallel(func(plan planner.Planner) error {
				plan.Task(task("A"))
				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(plan.Next(newStatuses())).To(EqualTasks([]task{"A"}))
			Expect(plan.State(newStatuses())).To(Equal(status.Unstarted))
		})

		It("returns no steps when the step completes", func() {
			plan, err := planner.NewParallel(func(plan planner.Planner) error {
				plan.Task(task("A"))
				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			statuses := []status.Type{status.Success, status.Failed, status.Running, status.Errored}
			for _, status := range statuses {
				state := newStatuses()
				Expect(state.Add(task("A"), status)).ToNot(HaveOccurred())

				Expect(plan.Next(state)).To(EqualTasks([]task{}))
				Expect(plan.State(state)).To(Equal(status))
			}
		})
	})

	Context("a serial plan with two steps", func() {
		It("returns the step for executor when it hasn't started yet", func() {
			plan, err := planner.NewSerial(func(plan planner.Planner) error {
				plan.Task(task("A"))
				plan.Task(task("B"))
				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(plan.Next(newStatuses())).To(EqualTasks([]task{"A"}))
			Expect(plan.State(newStatuses())).To(Equal(status.Unstarted))
		})

		It("returns it has started running when first step has passed", func() {
			plan, err := planner.NewSerial(func(plan planner.Planner) error {
				plan.Task(task("A"))
				plan.Task(task("B"))
				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			state := newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B"}))
			Expect(plan.State(state)).To(Equal(status.Running))
		})

		It("returns success when both step are successful", func() {
			plan, err := planner.NewSerial(func(plan planner.Planner) error {
				plan.Task(task("A"))
				plan.Task(task("B"))
				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			state := newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Success))
		})

		It("returns failed when first step has failed", func() {
			plan, err := planner.NewSerial(func(plan planner.Planner) error {
				plan.Task(task("A"))
				plan.Task(task("B"))
				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			state := newStatuses()
			Expect(state.Add(task("A"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Failed))
		})

		It("returns failed when second step has failed", func() {
			plan, err := planner.NewSerial(func(plan planner.Planner) error {
				plan.Task(task("A"))
				plan.Task(task("B"))
				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			state := newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Failed))
		})

		It("returns errored when first step has errored", func() {
			plan, err := planner.NewSerial(func(plan planner.Planner) error {
				plan.Task(task("A"))
				plan.Task(task("B"))
				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			state := newStatuses()
			Expect(state.Add(task("A"), status.Errored)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Errored))
		})
	})

	Context("a parallel plan with two steps", func() {
		It("returns the step for executor when it hasn't started yet", func() {
			plan, err := planner.NewParallel(func(plan planner.Planner) error {
				plan.Task(task("A"))
				plan.Task(task("B"))
				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(plan.Next(newStatuses())).To(EqualTasks([]task{"A", "B"}))
			Expect(plan.State(newStatuses())).To(Equal(status.Unstarted))

			state := newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Errored)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Errored))

			state = newStatuses()
			Expect(state.Add(task("B"), status.Errored)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Errored))

			state = newStatuses()
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"A"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("B"), status.Running)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"A"}))
			Expect(plan.State(state)).To(Equal(status.Running))
		})

		It("returns the aggregated state when both steps have finished", func() {
			plan, err := planner.NewParallel(func(plan planner.Planner) error {
				plan.Task(task("A"))
				plan.Task(task("B"))
				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			state := newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Success))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Failed))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Failed))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Errored)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Errored))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Errored)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Errored))
		})

		It("returns running if a single step is still running", func() {
			plan, err := planner.NewParallel(func(plan planner.Planner) error {
				plan.Task(task("A"))
				plan.Task(task("B"))
				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			state := newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Running)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Running))
		})
	})

	Context("with a composed serial and parallel plan", func() {
		var plan planner.Step

		BeforeEach(func() {
			plan, _ = planner.NewSerial(func(plan planner.Planner) error {
				err := plan.Parallel(func(plan planner.Planner) error {
					plan.Task(task("A"))
					plan.Task(task("B"))
					err := plan.Serial(func(plan planner.Planner) error {
						plan.Task(task("C"))
						plan.Task(task("D"))
						return nil
					})
					Expect(err).NotTo(HaveOccurred())

					err = plan.Parallel(func(plan planner.Planner) error {
						plan.Task(task("E"))
						err := plan.Serial(func(plan planner.Planner) error {
							plan.Task(task("F"))
							plan.Task(task("G"))
							return nil
						})
						Expect(err).NotTo(HaveOccurred())

						return nil
					})
					Expect(err).NotTo(HaveOccurred())

					return nil
				})
				Expect(err).NotTo(HaveOccurred())
				plan.Task(task("H"))
				return nil
			})
		})

		It("has an initial state", func() {
			Expect(plan.Next(newStatuses())).To(EqualTasks([]task{"A", "B", "C", "E", "F"}))
			Expect(plan.State(newStatuses())).To(Equal(status.Unstarted))
		})

		It("recommends no next steps on error", func() {
			state := newStatuses()
			Expect(state.Add(task("A"), status.Errored)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Errored))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("F"), status.Errored)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Errored))
		})

		It("recommends the next step on success of another", func() {
			state := newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B", "C", "E", "F"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("F"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B", "C", "E", "G"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("C"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("F"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B", "D", "E", "G"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("C"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("D"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("F"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B", "E", "G"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("C"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("D"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("F"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"E", "G"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("C"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("D"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("F"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("G"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"E"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("C"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("D"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("E"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("F"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("G"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"H"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("C"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("D"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("E"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("F"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("G"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("H"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Success))
		})

		It("recommends the correct steps on failure", func() {
			state := newStatuses()
			Expect(state.Add(task("A"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B", "C", "E", "F"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"C", "E", "F"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("C"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"E", "F"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("C"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("E"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("F"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Failed))
		})
	})

	When("a success step is defined", func() {
		It("only triggers when all serial steps are successful", func() {
			plan, err := planner.NewSerial(func(plan planner.Planner) error {
				plan.Task(task("A"))
				plan.Task(task("B"))
				err := plan.Success(func(plan planner.Planner) error {
					err := plan.Serial(func(plan planner.Planner) error {
						plan.Task(task("C"))
						return nil
					})
					Expect(err).NotTo(HaveOccurred())

					return nil
				})
				Expect(err).NotTo(HaveOccurred())

				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(plan.Next(newStatuses())).To(EqualTasks([]task{"A"}))
			Expect(plan.State(newStatuses())).To(Equal(status.Unstarted))

			state := newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"C"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Failed))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("C"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Success))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("C"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Failed))
		})

		It("only triggers when all parallel steps are successful", func() {
			plan, err := planner.NewParallel(func(plan planner.Planner) error {
				plan.Task(task("A"))
				plan.Task(task("B"))
				err := plan.Success(func(plan planner.Planner) error {
					err := plan.Serial(func(plan planner.Planner) error {
						plan.Task(task("C"))
						return nil
					})
					Expect(err).NotTo(HaveOccurred())

					return nil
				})
				Expect(err).NotTo(HaveOccurred())

				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(plan.Next(newStatuses())).To(EqualTasks([]task{"A", "B"}))
			Expect(plan.State(newStatuses())).To(Equal(status.Unstarted))

			state := newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"C"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Failed))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("C"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Success))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("C"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Failed))
		})
	})

	When("a failure step is defined", func() {
		It("only triggers when any serial steps have failed", func() {
			plan, err := planner.NewSerial(func(plan planner.Planner) error {
				plan.Task(task("A"))
				plan.Task(task("B"))
				err := plan.Failure(func(plan planner.Planner) error {
					err := plan.Serial(func(plan planner.Planner) error {
						plan.Task(task("C"))
						return nil
					})
					Expect(err).NotTo(HaveOccurred())

					return nil
				})
				Expect(err).NotTo(HaveOccurred())

				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(plan.Next(newStatuses())).To(EqualTasks([]task{"A"}))
			Expect(plan.State(newStatuses())).To(Equal(status.Unstarted))

			state := newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Success))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"C"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"C"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("C"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Failed))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("C"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Failed))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("C"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Failed))
		})

		It("only triggers when any serial steps have failed", func() {
			plan, err := planner.NewParallel(func(plan planner.Planner) error {
				plan.Task(task("A"))
				plan.Task(task("B"))
				err := plan.Failure(func(plan planner.Planner) error {
					err := plan.Serial(func(plan planner.Planner) error {
						plan.Task(task("C"))
						return nil
					})
					Expect(err).NotTo(HaveOccurred())

					return nil
				})
				Expect(err).NotTo(HaveOccurred())

				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(plan.Next(newStatuses())).To(EqualTasks([]task{"A", "B"}))
			Expect(plan.State(newStatuses())).To(Equal(status.Unstarted))

			state := newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Success))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"C"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"C"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("C"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Failed))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("C"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Failed))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("C"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Failed))
		})
	})

	When("a finally step is specified", func() {
		It("always runs no matter the state of serial", func() {
			plan, err := planner.NewSerial(func(plan planner.Planner) error {
				plan.Task(task("A"))
				err := plan.Finally(func(plan planner.Planner) error {
					err := plan.Serial(func(plan planner.Planner) error {
						plan.Task(task("B"))
						return nil

					})
					Expect(err).NotTo(HaveOccurred())

					return nil
				})
				Expect(err).NotTo(HaveOccurred())

				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(plan.Next(newStatuses())).To(EqualTasks([]task{"A"}))
			Expect(plan.State(newStatuses())).To(Equal(status.Unstarted))

			state := newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Success))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Failed))
		})

		It("always runs no matter the state of parallel", func() {
			plan, err := planner.NewParallel(func(plan planner.Planner) error {
				plan.Task(task("A"))
				err := plan.Finally(func(plan planner.Planner) error {
					err := plan.Serial(func(plan planner.Planner) error {
						plan.Task(task("B"))
						return nil
					})
					Expect(err).NotTo(HaveOccurred())

					return nil
				})
				Expect(err).NotTo(HaveOccurred())

				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(plan.Next(newStatuses())).To(EqualTasks([]task{"A"}))
			Expect(plan.State(newStatuses())).To(Equal(status.Unstarted))

			state := newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Success))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Failed))
		})
	})

	Context("the order precedence of success/failure and finally", func() {
		It("recommends success/failure before finally in serial", func() {
			plan, err := planner.NewSerial(func(plan planner.Planner) error {
				plan.Task(task("A"))
				err := plan.Success(func(plan planner.Planner) error {
					err := plan.Serial(func(plan planner.Planner) error {
						plan.Task(task("B"))
						return nil
					})
					Expect(err).NotTo(HaveOccurred())

					return nil
				})
				Expect(err).NotTo(HaveOccurred())

				err = plan.Failure(func(plan planner.Planner) error {
					err := plan.Serial(func(plan planner.Planner) error {
						plan.Task(task("C"))
						return nil
					})
					Expect(err).NotTo(HaveOccurred())

					return nil
				})
				Expect(err).NotTo(HaveOccurred())

				err = plan.Error(func(plan planner.Planner) error {
					err := plan.Serial(func(plan planner.Planner) error {
						plan.Task(task("D"))
						return nil
					})
					Expect(err).NotTo(HaveOccurred())

					return nil
				})
				Expect(err).NotTo(HaveOccurred())

				err = plan.Finally(func(plan planner.Planner) error {
					err := plan.Serial(func(plan planner.Planner) error {
						plan.Task(task("E"))
						return nil
					})
					Expect(err).NotTo(HaveOccurred())

					return nil
				})
				Expect(err).NotTo(HaveOccurred())

				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			state := newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"C"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"E"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("C"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"E"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Errored)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"D"}))
			Expect(plan.State(state)).To(Equal(status.Running))
		})

		It("recommends success/failure before finally in parallel", func() {
			plan, err := planner.NewParallel(func(plan planner.Planner) error {
				plan.Task(task("A"))
				err := plan.Success(func(plan planner.Planner) error {
					err := plan.Serial(func(plan planner.Planner) error {
						plan.Task(task("B"))
						return nil
					})
					Expect(err).NotTo(HaveOccurred())

					return nil
				})
				Expect(err).NotTo(HaveOccurred())

				err = plan.Failure(func(plan planner.Planner) error {
					err := plan.Serial(func(plan planner.Planner) error {
						plan.Task(task("C"))
						return nil
					})
					Expect(err).NotTo(HaveOccurred())

					return nil
				})
				Expect(err).NotTo(HaveOccurred())

				err = plan.Error(func(plan planner.Planner) error {
					err := plan.Serial(func(plan planner.Planner) error {
						plan.Task(task("D"))
						return nil
					})
					Expect(err).NotTo(HaveOccurred())

					return nil
				})
				Expect(err).NotTo(HaveOccurred())

				err = plan.Finally(func(plan planner.Planner) error {
					err := plan.Serial(func(plan planner.Planner) error {
						plan.Task(task("E"))
						return nil
					})
					Expect(err).NotTo(HaveOccurred())

					return nil
				})
				Expect(err).NotTo(HaveOccurred())

				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			state := newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"C"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"E"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("C"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"E"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Errored)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"D"}))
			Expect(plan.State(state)).To(Equal(status.Running))
		})
	})
	When("defining a try statement", func() {
		It("always return success", func() {
			plan, err := planner.NewSerial(func(plan planner.Planner) error {
				err := plan.Try(func(plan planner.Planner) error {
					err := plan.Serial(func(plan planner.Planner) error {
						plan.Task(task("A"))
						return nil
					})
					Expect(err).NotTo(HaveOccurred())

					return nil
				})
				Expect(err).NotTo(HaveOccurred())

				plan.Task(task("B"))
				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(plan.Next(newStatuses())).To(EqualTasks([]task{"A"}))
			Expect(plan.State(newStatuses())).To(Equal(status.Unstarted))

			state := newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Failed))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Success))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Success))
		})

		It("always return success", func() {
			plan, err := planner.NewParallel(func(plan planner.Planner) error {
				err := plan.Try(func(plan planner.Planner) error {
					err := plan.Parallel(func(plan planner.Planner) error {
						plan.Task(task("A"))
						return nil

					})
					Expect(err).NotTo(HaveOccurred())

					return nil
				})
				Expect(err).NotTo(HaveOccurred())

				plan.Task(task("B"))
				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(plan.Next(newStatuses())).To(EqualTasks([]task{"A", "B"}))
			Expect(plan.State(newStatuses())).To(Equal(status.Unstarted))

			state := newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Failed))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Success))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Success))
		})
	})

	When("handling number of attempts", func() {
		It("only reruns the tasks in serial", func() {
			plan, err := planner.NewSerial(func(plan planner.Planner) error {
				err := plan.Serial(func(plan planner.Planner) error {
					plan.Task(task("A"))
					plan.Task(task("B"))
					return nil
				}, planner.WithAttempts(2))
				Expect(err).NotTo(HaveOccurred())

				err = plan.Failure(func(plan planner.Planner) error {
					err := plan.Serial(func(plan planner.Planner) error {
						plan.Task(task("C"))
						return nil
					})
					Expect(err).NotTo(HaveOccurred())

					return nil
				})
				Expect(err).NotTo(HaveOccurred())

				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(plan.Next(newStatuses())).To(EqualTasks([]task{"A"}))
			Expect(plan.State(newStatuses())).To(Equal(status.Unstarted))

			state := newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Success))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"A"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Success))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("A"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"C"}))
			Expect(plan.State(state)).To(Equal(status.Running))
		})

		It("only reruns the tasks for parallel", func() {
			plan, err := planner.NewParallel(func(plan planner.Planner) error {
				err := plan.Parallel(func(plan planner.Planner) error {
					plan.Task(task("A"))
					plan.Task(task("B"))
					return nil
				}, planner.WithAttempts(2))
				Expect(err).NotTo(HaveOccurred())

				err = plan.Failure(func(plan planner.Planner) error {
					err := plan.Serial(func(plan planner.Planner) error {
						plan.Task(task("C"))
						return nil
					})
					Expect(err).NotTo(HaveOccurred())

					return nil
				})
				Expect(err).NotTo(HaveOccurred())

				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(plan.Next(newStatuses())).To(EqualTasks([]task{"A", "B"}))
			Expect(plan.State(newStatuses())).To(Equal(status.Unstarted))

			state := newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"A", "B"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"A", "B"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("A"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"C"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Failed)).ToNot(HaveOccurred())
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{}))
			Expect(plan.State(state)).To(Equal(status.Success))
		})
	})

	When("handling max steps in flight", func() {
		It("limits the number of steps returned at a time", func() {
			plan, _ := planner.NewParallel(func(plan planner.Planner) error {
				plan.Task(task("A"))
				plan.Task(task("B"))
				plan.Task(task("C"))
				return nil
			}, planner.WithMaxStepsInFlight(2))

			Expect(plan.Next(newStatuses())).To(EqualTasks([]task{"A", "B"}))
			Expect(plan.State(newStatuses())).To(Equal(status.Unstarted))

			state := newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B", "C"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"A", "C"}))
			Expect(plan.State(state)).To(Equal(status.Running))
		})

		It("limits deeply nested steps", func() {
			plan, err := planner.NewSerial(func(plan planner.Planner) error {
				err := plan.Parallel(func(plan planner.Planner) error {
					plan.Task(task("A"))
					plan.Task(task("B"))
					plan.Task(task("C"))
					return nil
				})
				Expect(err).NotTo(HaveOccurred())

				return nil
			}, planner.WithMaxStepsInFlight(2))
			Expect(err).NotTo(HaveOccurred())

			Expect(plan.Next(newStatuses())).To(EqualTasks([]task{"A", "B"}))
			Expect(plan.State(newStatuses())).To(Equal(status.Unstarted))

			state := newStatuses()
			Expect(state.Add(task("A"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"B", "C"}))
			Expect(plan.State(state)).To(Equal(status.Running))

			state = newStatuses()
			Expect(state.Add(task("B"), status.Success)).ToNot(HaveOccurred())

			Expect(plan.Next(state)).To(EqualTasks([]task{"A", "C"}))
			Expect(plan.State(state)).To(Equal(status.Running))
		})
	})
})
