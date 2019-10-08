package writers

import (
	"fmt"
	"io"

	"github.com/jtarchie/dothings/executor"
)

type Console struct {
	stdout executor.StringerWriter
	stderr executor.StringerWriter
}

var _ executor.Writer = &Console{}

func NewConsole(stdout executor.StringerWriter, stderr executor.StringerWriter) *Console {
	return &Console{
		stdout: stdout,
		stderr: stderr,
	}
}

func (c *Console) GetWriter(task executor.Tasker) (io.Writer, io.Writer) {
	_, _ = fmt.Fprintf(c.stderr, "initializing %s\n", task.ID())
	return c.stdout, c.stderr
}

func (c *Console) GetString(executor.Tasker) (string, string) {
	return c.stdout.String(), c.stderr.String()
}
