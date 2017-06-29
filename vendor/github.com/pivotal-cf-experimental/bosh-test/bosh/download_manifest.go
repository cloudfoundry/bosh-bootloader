package bosh

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type deploymentManifest struct {
	Manifest string `json:"manifest"`
}

func (c Client) DownloadManifest(deploymentName string) ([]byte, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/deployments/%s", c.config.URL, deploymentName), nil)
	if err != nil {
		return nil, err
	}

	request.SetBasicAuth(c.config.Username, c.config.Password)

	response, err := transport.RoundTrip(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		body, err := bodyReader(response.Body)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()

		return nil, fmt.Errorf("unexpected response %d %s:\n%s", response.StatusCode, http.StatusText(response.StatusCode), body)
	}

	var result deploymentManifest
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return []byte(result.Manifest), nil
}
