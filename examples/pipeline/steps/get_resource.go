package steps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/jtarchie/dothings/examples/pipeline/models"
	"github.com/jtarchie/dothings/planner"
	"github.com/jtarchie/dothings/status"
)

type GetResource struct {
	resource         *models.Resource
	versionManager   versionManager
	volumeManager    volumeManager
	containerManager containerManager
	params           map[string]interface{}
	timestamp        int64
}

func NewGetResource(
	r *models.Resource,
	versionManager versionManager,
	volumeManager volumeManager,
	containerManger containerManager,
	params map[string]interface{},
) *GetResource {
	return &GetResource{
		resource:         r,
		versionManager:   versionManager,
		volumeManager:    volumeManager,
		containerManager: containerManger,
		params:           params,
		timestamp:        time.Now().UnixNano(),
	}
}

func (g *GetResource) ID() string {
	return fmt.Sprintf("get resource: %s (%d)", g.resource.Name, g.timestamp)
}

func (g *GetResource) Execute(stdout io.Writer, stderr io.Writer) (status.Type, error) {
	resourcePath := g.volumeManager.Get(g.resource.Name, true)

	runner := g.containerManager

	workingDir := fmt.Sprintf("/tmp/build/get-%s", generateBuildGUID())
	runner.WorkingDir(workingDir)
	runner.Command("/opt/resource/in", workingDir)
	runner.Image(
		fmt.Sprintf("concourse/%s-resource", g.resource.Type),
		"latest",
	)
	runner.Volume(resourcePath, workingDir)
	runner.Privileged(true)

	payload := struct {
		Source  map[string]interface{} `json:"source,omitempty"`
		Version map[string]string      `json:"version,omitempty"`
		Params  map[string]interface{} `json:"params,omitempty"`
	}{
		Source:  g.resource.Source,
		Version: g.versionManager.GetLatestVersion(g.resource),
		Params:  g.params,
	}

	contents, err := json.Marshal(payload)
	if err != nil {
		return status.Failed, fmt.Errorf("get resource json marshal failed: %s", err)
	}

	err = runner.Run(
		bytes.NewBuffer(contents),
		stdout,
		stderr,
	)

	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return status.Failed, nil
		}
		return status.Errored, fmt.Errorf("get resource execute errored: %s", err)
	}

	return status.Success, nil
}

var _ planner.Tasker = &GetResource{}
