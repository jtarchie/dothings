package models

import (
	"reflect"
)

type input struct {
	Name string
}

type inputs []input

type output struct {
	Name string
}

type outputs []output

type task struct {
	Name   string `yaml:"task"`
	Config struct {
		Platform      string
		ImageResource struct {
			Type   string
			Source map[string]interface{}
		} `yaml:"image_resource"`
		Run struct {
			Path string
			Args []string
			Dir  string
		}
		Params  map[string]string
		Inputs  inputs
		Outputs outputs
	} `yaml:"config"`
	Image string
}

type get struct {
	Name string `yaml:"get"`
}

type stepParams map[string]interface{}

type put struct {
	Name      string `yaml:"put"`
	GetParams stepParams
}

type Step struct {
	Task       task       `yaml:",inline"`
	Get        get        `yaml:",inline"`
	Put        put        `yaml:",inline"`
	InParallel Steps      `yaml:"in_parallel"`
	Do         Steps      `yaml:"do"`
	Params     stepParams `yaml:"params"`
	Tags       []string
	Attempts   int
}

type Type int

func (t Type) String() string {
	switch t {
	case Get:
		return "Get"
	case Put:
		return "Put"
	case Task:
		return "Task"
	case InParallel:
		return "InParallel"
	case Do:
		return "Do"
	}
	return "Unknown"
}

const (
	Task Type = iota
	Get
	Put
	InParallel
	Do
	Unknown
)

func (step Step) Type() Type {
	if !reflect.DeepEqual(step.Task, task{}) {
		return Task
	}
	if !reflect.DeepEqual(step.Get, get{}) {
		return Get
	}
	if !reflect.DeepEqual(step.Put, put{}) {
		return Put
	}
	if step.InParallel != nil {
		return InParallel
	}
	if step.Do != nil {
		return Do
	}
	return Unknown
}

type Steps []Step

type Job struct {
	Name  string
	Steps Steps `yaml:"plan"`
}

type Jobs []Job

func (jobs Jobs) FindByName(name string) *Job {
	for _, job := range jobs {
		if job.Name == name {
			return &job
		}
	}
	return nil
}
