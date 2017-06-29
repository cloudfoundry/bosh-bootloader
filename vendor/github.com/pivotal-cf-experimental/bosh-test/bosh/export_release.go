package bosh

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type exportReleaseRequest struct {
	DeploymentName  string `json:"deployment_name"`
	ReleaseName     string `json:"release_name"`
	ReleaseVersion  string `json:"release_version"`
	StemcellOS      string `json:"stemcell_os"`
	StemcellVersion string `json:"stemcell_version"`
}

func (c Client) ExportRelease(deploymentName, releaseName, releaseVersion, stemcellOS, stemcellVersion string) (string, error) {
	content := exportReleaseRequest{
		DeploymentName:  deploymentName,
		ReleaseName:     releaseName,
		ReleaseVersion:  releaseVersion,
		StemcellOS:      stemcellOS,
		StemcellVersion: stemcellVersion,
	}

	requestBody, err := json.Marshal(content)
	if err != nil {
		return "", err
	}

	request, err := http.NewRequest("POST", fmt.Sprintf("%s/releases/export", c.config.URL), bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	request.SetBasicAuth(c.config.Username, c.config.Password)
	request.Header.Set("Content-Type", "application/json")

	response, err := transport.RoundTrip(request)
	if err != nil {
		return "", err
	}

	taskId, err := c.checkTaskStatus(response.Header.Get("Location"))
	if err != nil {
		return "", err
	}

	taskResult, err := c.TaskResult(taskId)
	if err != nil {
		return "", err
	}

	blobstoreID, ok := taskResult["blobstore_id"].(string)
	if !ok {
		return "", errors.New("could not find \"blobstore_id\" key in task result")
	}

	return blobstoreID, nil
}
