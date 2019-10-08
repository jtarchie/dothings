package main

import (
	"fmt"
	"github.com/jtarchie/dothings/executor"
	"github.com/jtarchie/dothings/executor/writers"
	"github.com/jtarchie/dothings/planner"
	"github.com/jtarchie/dothings/status"
	"github.com/jtarchie/dothings/tasks"
	"os"
)

func main() {
	stdout, stderr := writers.NewStringReaderRouter(os.Stdout), writers.NewStringReaderRouter(os.Stderr)

	plan, _ := planner.NewSerial(func(plan planner.Planner) error {
		for i := 0; i < 10; i++ {
			plan.Task(tasks.NewEcho(fmt.Sprintf("Hello, World %d", i), status.Success))
		}
		return nil
	})
	fmt.Println("Starting execution")
	executor.NewExecutor(plan, writers.NewConsole(stdout, stderr)).Wait()
	fmt.Println("Done executing")

	plan, _ = planner.NewParallel(func(plan planner.Planner) error {
		for i := 0; i < 10; i++ {
			plan.Task(tasks.NewEcho(fmt.Sprintf("Hello, World %d", i), status.Success))
		}
		return nil
	})
	fmt.Println("Starting execution")

	executor.NewExecutor(plan, writers.NewConsole(stdout, stderr)).Wait()

	fmt.Println("Done executing")
}
