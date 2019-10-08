package steps_test

import (
	"io/ioutil"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSteps(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Steps Suite")
}

var doc, _ = ioutil.ReadFile("../pipeline.yml")
