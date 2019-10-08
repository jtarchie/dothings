package writers_test

import (
	"io/ioutil"
	"net/http/httptest"

	"github.com/jtarchie/dothings/executor"
	"github.com/jtarchie/dothings/executor/writers"
	dothings "github.com/jtarchie/dothings/planner"
	"github.com/jtarchie/dothings/status"
	"github.com/jtarchie/dothings/tasks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("WebHandler", func() {
	It("writes a task stdout and stderr to a single buffer", func() {
		taskA := tasks.NewEcho("task 1", status.Success)
		taskB := tasks.NewEcho("task 2", status.Success)

		plan, _ := dothings.NewParallel(func(plan dothings.Planner) error {
			plan.Task(taskA)
			plan.Task(taskB)
			return nil
		})

		inMemory := writers.NewInMemory()
		statuses := status.NewStatuses()
		handler := writers.NewWebHandler(plan, inMemory, statuses)

		executor.NewExecutorWithStater(
			plan,
			inMemory,
			statuses,
		).Wait()

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		respBody, err := ioutil.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())

		body := string(respBody)
		Expect(body).To(ContainSubstring(`<div class="type-parallel">`))
		Expect(body).To(ContainSubstring(`<article class="card type-task status success">`))
		Expect(body).To(ContainSubstring(`<header class="id">task 1</header>`))
		Expect(body).To(ContainSubstring("out: executing task 1"))
		Expect(body).To(ContainSubstring("err: executing task 1"))
		Expect(body).To(ContainSubstring(`<header class="id">task 2</header>`))
		Expect(body).To(ContainSubstring("out: executing task 2"))
		Expect(body).To(ContainSubstring("err: executing task 2"))
	})
})
