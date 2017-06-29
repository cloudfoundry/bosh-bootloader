package bosh

import (
	"bytes"
	"fmt"
	"net/http"
)

func (c Client) Restart(deployment, job string, index int) error {
	request, err := http.NewRequest("PUT", fmt.Sprintf("%s/deployments/%s/jobs/%s/%d?state=restart", c.config.URL, deployment, job, index), bytes.NewBuffer([]byte{}))
	if err != nil {
		return err
	}

	request.SetBasicAuth(c.config.Username, c.config.Password)
	request.Header.Set("Content-Type", "text/yaml")
	response, err := transport.RoundTrip(request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusFound {
		responseBody, err := bodyReader(response.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("unexpected response %d %s:\n%s", response.StatusCode, http.StatusText(response.StatusCode), responseBody)
	}

	_, err = c.checkTaskStatus(response.Header.Get("Location"))
	if err != nil {
		return err
	}

	return nil
}
