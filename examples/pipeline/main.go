package main

import (
	"flag"
	"fmt"
	"github.com/jtarchie/dothings/examples/pipeline/steps/managers/docker"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/jtarchie/dothings/examples/pipeline/models"
	"github.com/jtarchie/dothings/examples/pipeline/steps"
	"github.com/jtarchie/dothings/executor"
	"github.com/jtarchie/dothings/executor/writers"
	"github.com/jtarchie/dothings/status"
	"gopkg.in/yaml.v2"
)

func main() {
	configFile := flag.String("config", "", "pipeline to configure")
	port := flag.Int("port", 8080, "port of the http server")
	flag.Parse()

	contents, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("could not read config file: %s", err)
	}

	pipeline := models.NewPipeline(models.ResourceTypes{
		models.ResourceType{Name: "registry-image", Type: "", Source: nil},
		models.ResourceType{Name: "docker-image", Type: "registry-image", Source: nil},
		models.ResourceType{Name: "git", Type: "registry-image", Source: nil},
	})
	err = yaml.UnmarshalStrict(contents, &pipeline)
	if err != nil {
		log.Fatalf("could not unmarshal pipeline from config file: %s", err)
	}

	builder := steps.NewBuilder(pipeline, docker.NewFactory())
	plan, err := builder.PlanForJob(pipeline.Jobs[0].Name)
	if err != nil {
		log.Fatalf("could not build plan for pipeline: %s", err)
	}

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
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
