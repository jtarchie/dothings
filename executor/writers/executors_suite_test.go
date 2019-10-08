package writers_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestExecutors(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Executor Writers Suite")
}
