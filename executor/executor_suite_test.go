package executor_test

import (
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/jtarchie/dothings/status"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestExecutor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Executor Suite")
}

type task string

func (i task) ID() string {
	return string(i)
}
func (i task) Execute(stdout io.Writer, _ io.Writer) (status.Type, error) {
	_, _ = fmt.Fprintf(stdout, "executed %s\n", string(i))
	return status.Success, nil
}

type erroringTask struct {
	task
}

func (i erroringTask) Execute(stdout io.Writer, stderr io.Writer) (status.Type, error) {
	_, _ = i.task.Execute(stdout, stderr)
	return -1, fmt.Errorf("error")
}

type failingTask struct {
	task
}

func (i failingTask) Execute(stdout io.Writer, stderr io.Writer) (status.Type, error) {
	_, _ = i.task.Execute(stdout, stderr)
	return status.Failed, nil
}

type timedTask struct {
	task
}

func (i timedTask) Execute(stdout io.Writer, stderr io.Writer) (status.Type, error) {
	duration, _ := time.ParseDuration(string(i.task))
	time.Sleep(duration)

	return i.task.Execute(stdout, stderr)
}

type blockingTask struct {
	wait    chan struct{}
	message string
}

func newBlockingTask(message string) *blockingTask {
	return &blockingTask{
		wait:    make(chan struct{}),
		message: message,
	}
}

func (i *blockingTask) ID() string {
	return i.message
}

func (i *blockingTask) Execute(stdout io.Writer, stderr io.Writer) (status.Type, error) {
	_, _ = fmt.Fprintf(stdout, "task %s\n", i.message)
	<-i.wait
	return status.Success, nil
}
