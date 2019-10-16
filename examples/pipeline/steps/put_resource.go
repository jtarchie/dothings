package steps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sort"
	"time"

	"github.com/jtarchie/dothings/examples/pipeline/models"
	"github.com/jtarchie/dothings/planner"
	"github.com/jtarchie/dothings/status"
)

type PutResource struct {
	resource         *models.Resource
	versionManager   versionManager
	volumeManager    VolumeManager
	containerManager ContainerManager
	params           map[string]interface{}
	timestamp        int64
}

func NewPutResource(
	r *models.Resource,
	versionManager versionManager,
	volumeManager VolumeManager,
	containerManager ContainerManager,
	params map[string]interface{},
) *PutResource {
	return &PutResource{
		resource:         r,
		versionManager:   versionManager,
		volumeManager:    volumeManager,
		containerManager: containerManager,
		params:           params,
		timestamp:        time.Now().UnixNano(),
	}
}

func (p *PutResource) ID() string {
	return fmt.Sprintf("put resource: %s (%d)", p.resource.Name, p.timestamp)
}

func (p *PutResource) Execute(stdout io.Writer, stderr io.Writer) (status.Type, error) {
	runner := p.containerManager

	workingDir := fmt.Sprintf("/tmp/build/put-%s", generateBuildGUID())
	runner.WorkingDir(workingDir)
	runner.Command("/opt/resource/out", workingDir)
	runner.Image(
		fmt.Sprintf("concourse/%s-resource", p.resource.Type),
		"latest",
	)

	names := []string{}
	volumes := p.volumeManager.All()
	for name := range volumes {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		path := volumes[name]
		runner.Volume(path, fmt.Sprintf("%s/%s", workingDir, name))
	}

	request := struct {
		Source map[string]interface{} `json:"source,omitempty"`
		Params map[string]interface{} `json:"params,omitempty"`
	}{
		Source: p.resource.Source,
		Params: p.params,
	}
	contents, err := json.Marshal(request)
	if err != nil {
		return status.Failed, fmt.Errorf("put resource json marshal failed: %s", err)
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

	response := struct {
		Version map[string]string `json:"version"`
	}{}
	if err = json.NewDecoder(responseBody).Decode(&response); err != nil {
		return status.Errored, fmt.Errorf("put resource response payload is invalid: %s", err)
	}

	p.versionManager.SetLatestVersion(p.resource, response.Version)

	return status.Success, nil
}

var _ planner.Tasker = &PutResource{}
