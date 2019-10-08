package planner

import "github.com/jtarchie/dothings/status"

type task struct {
	unitOfWork Tasker
}

var _ Step = task{}

func (t task) Tree() Tree {
	return Tree{
		node: Task,
		task: t.unitOfWork,
	}
}

func (t task) Next(currentState status.Stater, options ...stepOption) Tasks {
	s := &step{
		currentAttempt: 1,
	}
	for _, o := range options {
		o(s)
	}

	if states := currentState.Get(t.unitOfWork); len(states) >= s.currentAttempt {
		return Tasks{}
	}

	return Tasks{t.unitOfWork}
}

func (t task) State(currentState status.Stater, options ...stepOption) status.Type {
	s := &step{
		currentAttempt: 1,
	}
	for _, o := range options {
		o(s)
	}
	if states := currentState.Get(t.unitOfWork); len(states) >= s.currentAttempt {
		return states[s.currentAttempt-1]
	}
	if states := currentState.Get(t.unitOfWork); len(states) > 0 {
		return status.Running
	}

	return status.Unstarted
}
