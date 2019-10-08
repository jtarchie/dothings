package steps

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"io"
	"os/exec"
	"sort"
	"time"

	"github.com/jtarchie/dothings/examples/pipeline/models"

	"github.com/jtarchie/dothings/planner"
	"github.com/jtarchie/dothings/status"
)

type Task struct {
	step             models.Step
	volumeManager    volumeManager
	containerManager containerManager
	timestamp        int64
}

func NewTask(
	step models.Step,
	volumeManager volumeManager,
	containerManager containerManager,
) *Task {
	return &Task{
		step:             step,
		volumeManager:    volumeManager,
		containerManager: containerManager,
		timestamp:        time.Now().UnixNano(),
	}
}

func (t *Task) ID() string {
	return fmt.Sprintf("task: %s (%d)", t.step.Task.Name, t.timestamp)
}

func (t *Task) Execute(stdout io.Writer, stderr io.Writer) (status.Type, error) {
	workingPath := fmt.Sprintf("/tmp/build/%s", generateBuildGUID())
	runner := t.containerManager

	t.setupWorkingDirectory(runner, workingPath)
	t.setupInputs(workingPath)
	t.setupOutputs(workingPath)
	t.setupParams()
	err := t.setupImage()
	t.setupCommand()

	if err != nil {
		return status.Failed, err
	}

	err = runner.Run(
		nil,
		stdout,
		stderr,
	)

	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return status.Failed, nil
		}
		return status.Errored, fmt.Errorf("check resource execute errored: %s", err)
	}

	return status.Success, nil
}

func (t *Task) setupWorkingDirectory(runner containerManager, workingPath string) {
	if t.step.Task.Config.Run.Dir != "" {
		runner.WorkingDir(fmt.Sprintf("%s/%s", workingPath, t.step.Task.Config.Run.Dir))
	} else {
		runner.WorkingDir(workingPath)
	}
}

func (t *Task) setupImage() error {
	if image := t.step.Task.Image; image != "" {
		imagePath := t.volumeManager.Get(image, false)
		t.containerManager.ImageFromOCI(imagePath)
		return nil
	}

	source := t.step.Task.Config.ImageResource.Source
	repo, ok := source["repository"]
	if !ok {
		return fmt.Errorf("no repository defined for task's source: %v", source)
	}

	t.containerManager.Image(repo.(string), "latest")

	return nil
}

func (t *Task) setupParams() {
	paramNames := []string{}
	params := t.step.Task.Config.Params
	for name := range params {
		paramNames = append(paramNames, name)
	}
	sort.Strings(paramNames)
	for _, name := range paramNames {
		t.containerManager.EnvVar(name, params[name])
	}
}

func (t *Task) setupInputs(workingPath string) {
	for _, input := range t.step.Task.Config.Inputs {
		path := input.Name
		if input.Path != "" {
			path = input.Path
		}

		t.containerManager.Volume(
			t.volumeManager.Get(input.Name, false),
			fmt.Sprintf("%s/%s", workingPath, path),
		)
	}
}

func (t *Task) setupOutputs(workingPath string) {
	for _, output := range t.step.Task.Config.Outputs {
		path := output.Name
		if output.Path != "" {
			path = output.Path
		}

		t.containerManager.Volume(
			t.volumeManager.Get(output.Name, false),
			fmt.Sprintf("%s/%s", workingPath, path),
		)
	}
}

func (t *Task) setupCommand() {
	if t.step.Task.Config.Run.User != "" {
		t.containerManager.User(t.step.Task.Config.Run.User)
	}
	t.containerManager.Command(t.step.Task.Config.Run.Path, t.step.Task.Config.Run.Args...)
}

func generateBuildGUID() string {
	buffer := make([]byte, 10)
	_, _ = rand.Reader.Read(buffer)
	return base32.StdEncoding.EncodeToString(buffer)[0:6]
}

var _ planner.Tasker = &Task{}
