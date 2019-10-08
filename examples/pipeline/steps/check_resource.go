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

type CheckResource struct {
	resource         *models.Resource
	versionManager   versionManager
	timestamp        int64
	containerManager containerManager
}

func NewCheckResource(
	r *models.Resource,
	version versionManager,
	container containerManager,
) *CheckResource {
	return &CheckResource{
		resource:         r,
		versionManager:   version,
		containerManager: container,
		timestamp:        time.Now().UnixNano(),
	}
}

func (c *CheckResource) ID() string {
	return fmt.Sprintf("check resource: %s (%d)", c.resource.Name, c.timestamp)
}

func (c *CheckResource) Execute(stdout io.Writer, stderr io.Writer) (status.Type, error) {
	runner := c.containerManager

	workingDir := fmt.Sprintf("/tmp/build/check-%s", generateBuildGUID())
	runner.WorkingDir(workingDir)
	runner.Command("/opt/resource/check", workingDir)
	runner.Image(
		fmt.Sprintf("concourse/%s-resource", c.resource.Type),
		"latest",
	)

	request := struct {
		Source  map[string]interface{} `json:"source,omitempty"`
		Version map[string]string      `json:"version,omitempty"`
	}{
		Source:  c.resource.Source,
		Version: c.versionManager.GetLatestVersion(c.resource),
	}

	contents, err := json.Marshal(request)
	if err != nil {
		return status.Failed, fmt.Errorf("check resource json marshal failed: %s", err)
	}

	responseBody := &bytes.Buffer{}
	err = runner.Run(
		bytes.NewBuffer(contents),
		io.MultiWriter(stdout, responseBody),
		stderr,
	)

	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return status.Failed, nil
		}
		return status.Errored, fmt.Errorf("check resource execute errored: %s", err)
	}

	response := []map[string]string{}
	if err = json.NewDecoder(responseBody).Decode(&response); err != nil {
		return status.Errored, fmt.Errorf("check resource response payload is invalid: %s", err)
	}

	if len(response) > 0 {
		latestVersion := response[0]
		c.versionManager.SetLatestVersion(c.resource, latestVersion)
	}

	return status.Success, nil
}

var _ planner.Tasker = &CheckResource{}
