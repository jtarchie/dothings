package planner

import "github.com/jtarchie/dothings/status"

type try struct{ *plan }

var _ Step = &try{}

func (p *try) Tree() Tree {
	return p.tree(Try)
}

func (p *try) State(currentState status.Stater, options ...stepOption) status.Type {
	s := &step{
		currentAttempt: 1,
	}
	for _, o := range options {
		o(s)
	}
	state := p.steps[0].State(currentState, withCurrentAttempt(s.currentAttempt))
	if state == status.Failed {
		return status.Success
	}

	return state
}
