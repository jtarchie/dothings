package executor

import (
	"fmt"
	"io"
	"log"
	"runtime"
	"time"

	"github.com/jtarchie/dothings/planner"
	"github.com/jtarchie/dothings/status"
)

type Tasker interface {
	ID() string
	Execute(io.Writer, io.Writer) (status.Type, error)
}

type Writer interface {
	GetWriter(Tasker) (io.Writer, io.Writer)
	GetString(Tasker) (string, string)
}

type StringerWriter interface {
	fmt.Stringer
	io.Writer
}

type Executor struct {
	plan   planner.Step
	writer Writer
	stater status.Stater
}

func (e *Executor) Wait() status.Type {
	queue := make(chan Tasker)
	statuses := e.stater

	go func() {
		for task := range queue {
			go func(task Tasker) {
				stdout, stderr := e.writer.GetWriter(task)

				err := statuses.Add(task, status.Running)
				if err != nil {
					log.Printf("could not start task %s to state Running", task.ID())
					return
				}
				finalState, err := task.Execute(stdout, stderr)
				if err != nil {
					log.Printf("task failed execution: %s", err)
					finalState = status.Errored
				}
				err = statuses.Add(task, finalState)
				if err != nil {
					log.Printf("could not finished task %s to state %d", task.ID(), finalState)
				}
			}(task)
			runtime.Gosched()
		}
	}()

	for {
		tasks := e.plan.Next(statuses)

		if 0 < len(tasks) {
			for _, task := range tasks {
				err := statuses.Add(task, status.Unstarted)
				if err != nil {
					log.Printf("could not queue task %s to state Unstarted", task.ID())
					continue
				}
				queue <- task
			}
		}

		if len(tasks) == 0 {
			v := e.plan.State(statuses)
			switch v {
			case status.Running, status.Unstarted:
				break
			default:
				close(queue)
				return v
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func NewExecutor(
	plan planner.Step,
	writer Writer,
) *Executor {
	return &Executor{
		plan:   plan,
		writer: writer,
		stater: status.NewStatuses(),
	}
}

func NewExecutorWithStater(
	plan planner.Step,
	writer Writer,
	stater status.Stater,
) *Executor {
	return &Executor{
		plan:   plan,
		writer: writer,
		stater: stater,
	}
}
