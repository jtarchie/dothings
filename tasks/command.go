package tasks

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/jtarchie/dothings/planner"

	"github.com/jtarchie/dothings/status"
)

type LocalCommand struct {
	command string
	args    []string
}

var _ planner.Tasker = &LocalCommand{}

func NewCommand(command string, args ...string) *LocalCommand {
	return &LocalCommand{
		command: command,
		args:    args,
	}
}

func (c *LocalCommand) ID() string {
	return fmt.Sprintf("command: %s %s", c.command, strings.Join(c.args, " "))
}

func (c *LocalCommand) Execute(stdout io.Writer, stderr io.Writer) (status.Type, error) {
	command := exec.Command(c.command, c.args...)
	command.Stdout = stdout
	command.Stderr = stderr
	err := command.Run()

	if err != nil {
		return status.Failed, nil
	}

	return status.Success, nil
}
