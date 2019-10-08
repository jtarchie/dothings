package planner

import (
	"fmt"
	"io"
	"sort"

	"github.com/jtarchie/dothings/status"
)

type Tasks []Tasker

func (a Tasks) Len() int           { return len(a) }
func (a Tasks) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Tasks) Less(i, j int) bool { return a[i].ID() < a[j].ID() }

type stepOption func(*step)

type step struct {
	currentAttempt int
}

func withCurrentAttempt(currentAttempt int) func(*step) {
	return func(s *step) {
		s.currentAttempt = currentAttempt
	}
}

type Tree struct {
	node     planType
	task     Tasker
	children []Tree
}

func (t Tree) Type() planType {
	return t.node
}

func (t Tree) Children() []Tree {
	return t.children
}

func (t Tree) Task() Tasker {
	return t.task
}

type Step interface {
	Next(status.Stater, ...stepOption) Tasks
	State(status.Stater, ...stepOption) status.Type
	Tree() Tree
}

type configOption func(*plan)

func WithAttempts(attempts int) func(p *plan) {
	return func(p *plan) {
		p.attempts = attempts
	}
}

func WithMaxStepsInFlight(steps int) func(p *plan) {
	return func(p *plan) {
		p.maxSteps = steps
	}
}

type Planner interface {
	Task(Tasker) task
	Parallel(func(Planner) error, ...configOption) error
	Serial(func(Planner) error, ...configOption) error
	Success(func(Planner) error) error
	Failure(func(Planner) error) error
	Finally(func(Planner) error) error
	Try(func(Planner) error) error
	Error(func(plan Planner) error) error
}

type plan struct {
	steps   []Step
	success *success
	failure *failure
	finally *finally
	errored *errored

	attempts int
	maxSteps int
}

var _ Planner = &plan{}
var _ Step = &plan{}

func (p *plan) Tree() Tree {
	return Tree{}
}

func (p *plan) tree(name planType) Tree {
	nodes := []Tree{}
	for _, step := range p.steps {
		nodes = append(nodes, step.Tree())
	}
	if p.success != nil {
		nodes = append(nodes, p.success.Tree())
	}
	if p.failure != nil {
		nodes = append(nodes, p.failure.Tree())
	}
	if p.finally != nil {
		nodes = append(nodes, p.finally.Tree())
	}
	return Tree{
		node:     name,
		children: nodes,
	}
}

type Tasker interface {
	ID() string
	Execute(io.Writer, io.Writer) (status.Type, error)
}

func (p *plan) Task(unit Tasker) task {
	t := task{unit}
	p.steps = append(p.steps, t)
	return t
}

func (p *plan) Next(currentState status.Stater, options ...stepOption) Tasks {
	s := &step{
		currentAttempt: 1,
	}
	for _, o := range options {
		o(s)
	}
	tasks := p.steps[0].Next(currentState, withCurrentAttempt(s.currentAttempt))
	sort.Sort(tasks)
	return tasks
}

func (p *plan) State(currentState status.Stater, options ...stepOption) status.Type {
	s := &step{
		currentAttempt: 1,
	}
	for _, o := range options {
		o(s)
	}
	return p.steps[0].State(currentState, withCurrentAttempt(s.currentAttempt))
}

type success struct{ *plan }

func (p *success) Tree() Tree {
	return p.tree(Success)
}

type failure struct{ *plan }

func (p *failure) Tree() Tree {
	return p.tree(Failure)
}

type errored struct{ *plan }

func (p *errored) Tree() Tree {
	return p.tree(Failure)
}

type finally struct{ *plan }

func (p *finally) Tree() Tree {
	return p.tree(Finally)
}

func (p *plan) Success(fun func(Planner) error) error {
	plan := &success{newPlan()}
	err := fun(plan)
	if err != nil {
		return fmt.Errorf("could not create success step: %s", err)
	}
	p.success = plan
	return nil
}

func (p *plan) Failure(fun func(Planner) error) error {
	plan := &failure{newPlan()}
	err := fun(plan)
	if err != nil {
		return fmt.Errorf("could not create failure step: %s", err)
	}
	p.failure = plan
	return nil
}

func (p *plan) Error(fun func(plan Planner) error) error {
	plan := &errored{newPlan()}
	err := fun(plan)
	if err != nil {
		return fmt.Errorf("could not create error step: %s", err)
	}
	p.errored = plan
	return nil
}

func (p *plan) Finally(fun func(Planner) error) error {
	plan := &finally{newPlan()}
	err := fun(plan)
	if err != nil {
		return fmt.Errorf("could not create finally step: %s", err)
	}
	p.finally = plan
	return nil
}

func (p *plan) Parallel(fun func(Planner) error, options ...configOption) error {
	plan := &parallel{newPlan()}

	for _, o := range options {
		o(plan.plan)
	}

	err := fun(plan)
	if err != nil {
		return fmt.Errorf("could not create parallel step: %s", err)
	}
	p.steps = append(p.steps, plan)
	return nil
}

func (p *plan) Serial(fun func(Planner) error, options ...configOption) error {
	plan := &serial{newPlan()}

	for _, o := range options {
		o(plan.plan)
	}

	err := fun(plan)
	if err != nil {
		return fmt.Errorf("could not create serial step: %s", err)
	}
	p.steps = append(p.steps, plan)
	return nil
}

func (p *plan) Try(fun func(Planner) error) error {
	plan := &try{newPlan()}
	err := fun(plan)
	if err != nil {
		return fmt.Errorf("could not create try step: %s", err)
	}
	p.steps = append(p.steps, plan)
	return nil
}

func newPlan() *plan {
	return &plan{
		attempts: 1,
	}
}
