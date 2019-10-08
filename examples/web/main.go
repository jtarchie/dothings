package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jtarchie/dothings/executor"

	"github.com/jtarchie/dothings/executor/writers"
	"github.com/jtarchie/dothings/planner"
	"github.com/jtarchie/dothings/status"
)

type dotter struct {
	duration time.Duration
	polling  time.Duration
}

func (d *dotter) ID() string {
	return fmt.Sprintf("dotter %s", d.duration)
}

func (d *dotter) Execute(stdout io.Writer, _ io.Writer) (status.Type, error) {
	duration := d.duration
	for range time.Tick(d.polling) {
		_, _ = fmt.Fprintf(stdout, "%s remaining\n", duration)
		duration = duration - d.polling
		if duration <= 0 {
			break
		}
	}
	return status.Success, nil
}

var _ executor.Tasker = &dotter{}

func main() {
	pollingIntervalStr := flag.String("polling-interval", "5s", "the duration for the")
	durationStr := flag.String("duration", "10s", "the duration for the ")
	numTasks := flag.Int("num-tasks", 10, "the number of tasks to run")
	port := flag.Int("port", 8080, "port of the http server")

	flag.Parse()

	pollingInterval, err := time.ParseDuration(*pollingIntervalStr)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	duration, err := time.ParseDuration(*durationStr)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	log.Println("configuration")
	log.Printf("polling: %s", pollingInterval)
	log.Printf("duration: %s", duration)
	log.Printf("tasks: %d", *numTasks)
	log.Printf("port: %d", *port)

	plan, _ := planner.NewParallel(func(plan planner.Planner) error {
		for i := 1; i <= *numTasks; i++ {
			plan.Task(&dotter{
				polling:  pollingInterval,
				duration: time.Duration(i) * duration,
			})
		}
		return nil
	})

	log.Println("starting execution")
	inMemory := writers.NewInMemory()
	statuses := status.NewStatuses()
	handler := writers.NewWebHandler(plan, inMemory, statuses)

	http.Handle("/", handler)

	go executor.NewExecutorWithStater(
		plan,
		inMemory,
		statuses,
	).Wait()

	log.Printf("listening on http://localhost:%d", *port)
	log.Printf("current status: %s", plan.State(statuses))
	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
	}()

	for {
		log.Printf("current status: %s", plan.State(statuses))
		if plan.State(statuses) == status.Success {
			log.Println("successfully finished, exiting")
			os.Exit(0)
		}
		time.Sleep(pollingInterval)
	}
}
