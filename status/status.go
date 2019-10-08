package status

import (
	"fmt"
	"sync"
)

type Type int

const (
	Unstarted Type = iota
	Running
	Success
	Failed
	Errored
)

func (t Type) String() string {
	switch t {
	case Unstarted:
		return "unstarted"
	case Running:
		return "running"
	case Success:
		return "success"
	case Failed:
		return "failed"
	case Errored:
		return "errored"
	}
	return ""
}

type Identifier interface {
	ID() string
}

type currentState struct {
	sync.Mutex
	values map[string][]Type
}

type Stater interface {
	Get(task Identifier) []Type
	Add(task Identifier, s Type) error
}

func NewStatuses() Stater {
	return &currentState{
		values: map[string][]Type{},
	}
}

func (c *currentState) Add(task Identifier, s Type) error {
	c.Lock()
	defer c.Unlock()
	statuses, ok := c.values[task.ID()]

	if !ok {
		if s == Unstarted {
			c.values[task.ID()] = append(c.values[task.ID()], s)
			return nil
		}
		return fmt.Errorf("the set status %s cannot be an initial Stater", s)
	}

	current := statuses[len(statuses)-1]
	if current == Unstarted && s == Running {
		c.values[task.ID()][len(statuses)-1] = s
		return nil
	}
	if current == Running && finalState(s) {
		c.values[task.ID()][len(statuses)-1] = s
		return nil
	}
	if finalState(current) && s == Unstarted {
		c.values[task.ID()] = append(c.values[task.ID()], s)
		return nil
	}

	return fmt.Errorf("cannot transition from %s to %s", current, s)
}

func finalState(s Type) bool {
	return s == Failed || s == Success || s == Errored
}

func (c *currentState) Get(task Identifier) []Type {
	c.Lock()
	defer c.Unlock()

	if statuses, ok := c.values[task.ID()]; ok {
		tmp := make([]Type, len(statuses))
		copy(tmp, statuses)
		return tmp
	}
	return []Type{}
}
