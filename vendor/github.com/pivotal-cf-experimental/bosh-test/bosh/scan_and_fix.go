package bosh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	yaml "gopkg.in/yaml.v2"
)

func (c Client) ScanAndFix(deploymentName, jobName string, jobIndices []int) error {
	return c.doScanAndFixRequest(deploymentName, map[string]interface{}{
		"jobs": map[string][]int{
			jobName: jobIndices,
		},
	})
}

func (c Client) ScanAndFixAll(manifestYAML []byte) error {
	var manifest struct {
		Name string
		Jobs []struct {
			Name      string
			Instances int
		}
	}
	err := yaml.Unmarshal(manifestYAML, &manifest)
	if err != nil {
		return err
	}

	jobs := make(map[string][]int)
	for _, j := range manifest.Jobs {
		if j.Instances > 0 {
			var indices []int
			for i := 0; i < j.Instances; i++ {
				indices = append(indices, i)
			}
			jobs[j.Name] = indices
		}
	}

	return c.doScanAndFixRequest(manifest.Name, map[string]interface{}{
		"jobs": jobs,
	})
}

func (c Client) doScanAndFixRequest(deploymentName string, payload map[string]interface{}) error {
	requestBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	request, err := http.NewRequest("PUT", fmt.Sprintf("%s/deployments/%s/scan_and_fix", c.config.URL, deploymentName), bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	request.SetBasicAuth(c.config.Username, c.config.Password)
	request.Header.Set("Content-Type", "application/json")

	response, err := transport.RoundTrip(request)
	if err != nil {
		return err
	}

	responseBody, err := bodyReader(response.Body)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusFound {
		return fmt.Errorf("unexpected response %d %s:\n%s", response.StatusCode, http.StatusText(response.StatusCode), responseBody)
	}

	_, err = c.checkTaskStatus(response.Header.Get("Location"))
	if err != nil {
		return err
	}

	return nil
}
