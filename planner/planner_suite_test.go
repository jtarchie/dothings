package planner_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/jtarchie/dothings/status"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPlanner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Planner Suite")
}

type task string

func (i task) ID() string {
	return string(i)
}
func (i task) Execute(stdout io.Writer, _ io.Writer) (status.Type, error) {
	_, _ = fmt.Fprintf(stdout, "executed %s\n", string(i))
	return status.Success, nil
}
