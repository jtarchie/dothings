package tasks

import (
	"fmt"
	"io"

	"github.com/jtarchie/dothings/planner"

	"github.com/jtarchie/dothings/status"
)

type Echo struct {
	message string
	status  status.Type
}

var _ planner.Tasker = &Echo{}

func NewEcho(message string, status status.Type) *Echo {
	return &Echo{
		message: message,
		status:  status,
	}
}

func (e *Echo) ID() string {
	return e.message
}

func (e *Echo) Execute(stdout io.Writer, stderr io.Writer) (status.Type, error) {
	fmt.Fprintf(stdout, "out: executing %s\n", e.message)
	fmt.Fprintf(stderr, "err: executing %s\n", e.message)
	return e.status, nil
}
