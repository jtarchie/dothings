package planner

import "github.com/jtarchie/dothings/status"

type serial struct {
	*plan
}

var _ Step = &serial{}

func (p *serial) Tree() Tree {
	return p.tree(Serial)
}

func (p *serial) Next(currentState status.Stater, options ...stepOption) Tasks {
	s := &step{
		currentAttempt: 1,
	}
	for _, o := range options {
		o(s)
	}

	names, failed := Tasks{}, false
	for currentAttempt := 1; currentAttempt <= p.attempts; currentAttempt++ {
		for _, step := range p.steps {
			switch step.State(currentState, withCurrentAttempt(currentAttempt)) {
			case status.Success:
				continue
			case status.Failed, status.Errored:
				names = Tasks{}
				failed = true
				goto outOfStepLoop
			default:
				n := step.Next(currentState, withCurrentAttempt(currentAttempt))
				if p.maxSteps > 0 && len(n) >= p.maxSteps {
					return n[0:p.maxSteps]
				}
				return n
			}
		}
	outOfStepLoop:
		s.currentAttempt = currentAttempt
		if len(names) == 0 && !failed {
			break
		}
		if len(names) > 0 {
			break
		}
	}

	currentStatus := p.status(currentState, s.currentAttempt)

	if p.plan.success != nil && currentStatus == status.Success {
		if n := p.plan.success.Next(currentState, withCurrentAttempt(s.currentAttempt)); len(n) > 0 {
			names = append(names, n...)
		}
	}

	if p.plan.failure != nil && currentStatus == status.Failed {
		if n := p.plan.failure.Next(currentState, withCurrentAttempt(s.currentAttempt)); len(n) > 0 {
			names = append(names, n...)
		}
	}

	if p.plan.errored != nil && currentStatus == status.Errored {
		if n := p.plan.errored.Next(currentState, withCurrentAttempt(s.currentAttempt)); len(n) > 0 {
			names = append(names, n...)
		}
	}

	if p.plan.finally != nil {
		if n := p.plan.finally.Next(currentState, withCurrentAttempt(s.currentAttempt)); len(n) > 0 {
			names = append(names, n...)
		}
	}

	if len(names) > 0 {
		return Tasks{names[0]}
	}
	return Tasks{}
}

func (p *serial) State(currentState status.Stater, options ...stepOption) status.Type {
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

func (p *serial) status(currentState status.Stater, _ int) status.Type {
	statuses := map[status.Type]int{}

	for currentAttempt := 1; currentAttempt <= p.attempts; currentAttempt++ {
		statuses = map[status.Type]int{}
		for _, step := range p.steps {
			statuses[step.State(currentState, withCurrentAttempt(currentAttempt))]++
		}
		if len(statuses) == 1 {
			for status := range statuses {
				return status
			}
		}
	}

	if _, ok := statuses[status.Errored]; ok {
		return status.Errored
	}

	if _, ok := statuses[status.Failed]; ok {
		return status.Failed
	}

	return status.Running
}

func NewSerial(fun func(Planner) error, options ...configOption) (Step, error) {
	plan := &serial{newPlan()}

	for _, o := range options {
		o(plan.plan)
	}

	err := fun(plan)
	if err != nil {
		return nil, err
	}

	return plan, nil
}
