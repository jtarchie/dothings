package planner

import (
	"sort"

	"github.com/jtarchie/dothings/status"
)

type parallel struct {
	*plan
}

var _ Step = &parallel{}

func (p *parallel) Tree() Tree {
	return p.tree(Parallel)
}

func (p *parallel) Next(currentState status.Stater, options ...stepOption) Tasks {
	s := &step{
		currentAttempt: 1,
	}
	for _, o := range options {
		o(s)
	}

	names := Tasks{}
	for currentAttempt := 1; currentAttempt <= p.attempts; currentAttempt++ {
		for _, step := range p.steps {
			names = append(names, step.Next(currentState, withCurrentAttempt(currentAttempt))...)
		}

		s.currentAttempt = currentAttempt
		if len(names) > 0 {
			break
		}
	}

	currentStatus := p.status(currentState, s.currentAttempt)

	if p.plan.success != nil && currentStatus == status.Success && len(names) == 0 {
		names = append(names, p.plan.success.Next(currentState, withCurrentAttempt(s.currentAttempt))...)
	}

	if p.plan.failure != nil && currentStatus == status.Failed && len(names) == 0 {
		names = append(names, p.plan.failure.Next(currentState, withCurrentAttempt(s.currentAttempt))...)
	}

	if p.plan.errored != nil && currentStatus == status.Errored && len(names) == 0 {
		names = append(names, p.plan.errored.Next(currentState, withCurrentAttempt(s.currentAttempt))...)
	} else if currentStatus == status.Errored {
		return Tasks{}
	}

	if p.plan.finally != nil && len(names) == 0 {
		names = append(names, p.plan.finally.Next(currentState, withCurrentAttempt(s.currentAttempt))...)
	}

	if p.maxSteps > 0 && len(names) >= p.maxSteps {
		return names[0:p.maxSteps]
	}

	sort.Sort(names)
	return names
}

func (p *parallel) State(currentState status.Stater, options ...stepOption) status.Type {
	s := &step{
		currentAttempt: 1,
	}
	for _, o := range options {
		o(s)
	}
	statuses := map[status.Type]int{}

	statuses[p.status(currentState, s.currentAttempt)]++

	if p.plan.success != nil {
		statuses[p.plan.success.State(currentState, withCurrentAttempt(s.currentAttempt))]++
	}

	if p.plan.finally != nil {
		statuses[p.plan.finally.State(currentState, withCurrentAttempt(s.currentAttempt))]++
	}

	if len(statuses) == 1 {
		if _, failed := statuses[status.Failed]; failed && p.failure != nil && p.failure.State(currentState, withCurrentAttempt(s.currentAttempt)) <= status.Running {
			return status.Running
		}
		for status := range statuses {
			return status
		}
	}

	if _, ok := statuses[status.Failed]; ok {
		if p.finally != nil && p.finally.State(currentState, withCurrentAttempt(s.currentAttempt)) <= status.Running {
			return status.Running
		}

		return status.Failed
	}

	return status.Running
}

func (p *parallel) status(currentState status.Stater, _ int) status.Type {
	statuses := map[status.Type]int{}

	for currentAttempt := 1; currentAttempt <= p.attempts; currentAttempt++ {
		statuses = map[status.Type]int{}
		for _, step := range p.steps {
			statuses[step.State(currentState, withCurrentAttempt(currentAttempt))]++
		}

		if _, ok := statuses[status.Unstarted]; ok {
			break
		}
		if _, ok := statuses[status.Running]; ok {
			break
		}
	}

	if len(statuses) == 1 {
		for status := range statuses {
			return status
		}
	}

	if _, ok := statuses[status.Errored]; ok {
		return status.Errored
	}

	if _, ok := statuses[status.Unstarted]; ok {
		return status.Running
	}

	if _, ok := statuses[status.Running]; ok {
		return status.Running
	}

	if _, ok := statuses[status.Failed]; ok {
		return status.Failed
	}

	return status.Running
}

func NewParallel(fun func(Planner) error, options ...configOption) (Step, error) {
	plan := &parallel{newPlan()}

	for _, o := range options {
		o(plan.plan)
	}

	err := fun(plan)
	if err != nil {
		return nil, err
	}

	return plan, nil
}
