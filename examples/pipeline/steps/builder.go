package steps

import (
	"fmt"
	"github.com/jtarchie/dothings/examples/pipeline/steps/managers/docker"

	"github.com/jtarchie/dothings/examples/pipeline/models"
	"github.com/jtarchie/dothings/examples/pipeline/steps/managers"
	"github.com/jtarchie/dothings/planner"
)

type builder struct {
	pipeline       models.Pipeline
	versionManager versionManager
	volumeManager  volumeManager
	//containerManager containerManager
}

func NewBuilder(pipeline models.Pipeline) *builder {
	return &builder{
		pipeline:       pipeline,
		versionManager: managers.NewResourceVersionManager(),
		volumeManager:  managers.NewResourceVolumeManager(docker.DefaultExecutor),
		//containerManager: managers.NewDockerManager(),
	}
}

func (b *builder) PlanForJob(jobName string) (planner.Step, error) {
	job := b.pipeline.Jobs.FindByName(jobName)

	if job == nil {
		return nil, fmt.Errorf("job '%s' not found", jobName)
	}

	return planner.NewSerial(func(plan planner.Planner) error {
		err := b.createPlanFromSteps(plan, job.Steps)
		if err != nil {
			return fmt.Errorf("with job '%s': %s", jobName, err)
		}
		return nil
	})
}

func (b *builder) createPlanFromSteps(plan planner.Planner, steps models.Steps) error {
	for _, step := range steps {
		switch step.Type() {
		case models.Get:
			err := b.setupGet(step, plan)
			if err != nil {
				return err
			}
		case models.Task:
			b.setupTask(plan, step)
		case models.Put:
			err := b.setupPut(step, plan)
			if err != nil {
				return err
			}
		case models.InParallel:
			err := plan.Parallel(func(plan planner.Planner) error {
				return b.createPlanFromSteps(plan, step.InParallel)
			})
			if err != nil {
				return fmt.Errorf("in_parallel not buildable: %s", err)
			}
		case models.Do:
			err := plan.Serial(func(plan planner.Planner) error {
				return b.createPlanFromSteps(plan, step.Do)
			})
			if err != nil {
				return fmt.Errorf("do not buildable: %s", err)
			}
		default:
			return fmt.Errorf("step type '%s' not supported", step.Type())
		}
	}
	return nil
}

func (b *builder) setupPut(step models.Step, plan planner.Planner) error {
	resourceName := step.Put.Name
	resource := b.pipeline.Resources.FindByName(resourceName)
	if resource == nil {
		return fmt.Errorf("resource '%s' not found for put", resourceName)
	}
	plan.Serial(func(plan planner.Planner) error {
		plan.Task(NewPutResource(
			resource,
			b.versionManager,
			b.volumeManager,
			managers.NewDockerManager(docker.DefaultExecutor),
			step.Params,
		))
		plan.Task(NewGetResource(
			resource,
			b.versionManager,
			b.volumeManager,
			managers.NewDockerManager(docker.DefaultExecutor),
			step.Put.GetParams,
		))
		return nil
	})
	return nil
}

func (b *builder) setupTask(plan planner.Planner, step models.Step) {
	plan.Task(NewTask(
		step,
		b.volumeManager,
		managers.NewDockerManager(docker.DefaultExecutor),
	))
}

func (b *builder) setupGet(step models.Step, plan planner.Planner) error {
	resourceName := step.Get.Name
	resource := b.pipeline.Resources.FindByName(resourceName)
	if resource == nil {
		return fmt.Errorf("resource '%s' not found for get", resourceName)
	}
	plan.Serial(func(plan planner.Planner) error {
		plan.Task(NewCheckResource(
			resource,
			b.versionManager,
			managers.NewDockerManager(docker.DefaultExecutor),
		))
		plan.Task(NewGetResource(
			resource,
			b.versionManager,
			b.volumeManager,
			managers.NewDockerManager(docker.DefaultExecutor),
			step.Params,
		))
		return nil
	})
	return nil
}
