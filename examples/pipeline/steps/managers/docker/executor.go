package docker

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . CommandExecutor
type CommandExecutor interface {
	Run(
		io.Reader,
		io.Writer,
		io.Writer,
		string,
		...string,
	) error
}

type Executor func(
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	command string,
	args ...string,
) error

func (r Executor) Run(
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	command string,
	args ...string,
) error {
	return r(stdin, stdout, stderr, command, args...)
}

func defaultExecutor(
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	command string,
	args ...string,
) error {
	if stderr != nil {
		stderr.Write([]byte(fmt.Sprintf("$ %s %s\n", command, strings.Join(args, " "))))
	}
	c := exec.Command(command, args...)
	c.Stdin = stdin
	c.Stdout = stdout
	c.Stderr = stderr

	return c.Run()
}

var DefaultExecutor = Executor(defaultExecutor)
