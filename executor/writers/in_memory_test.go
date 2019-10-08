package writers_test

import (
	"github.com/jtarchie/dothings/executor/writers"
	"github.com/jtarchie/dothings/status"
	"github.com/jtarchie/dothings/tasks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InMemory", func() {
	It("can assign and read back ", func() {
		task := tasks.NewEcho("task 1", status.Success)
		writer := writers.NewInMemory()
		stdout, stderr := writer.GetWriter(task)

		Expect(stdout).To(Equal(stderr))
		_, err := stdout.Write([]byte("hello world"))
		Expect(err).NotTo(HaveOccurred())

		str, _ := writer.GetString(task)
		Expect(str).To(Equal("hello world"))
	})
})
