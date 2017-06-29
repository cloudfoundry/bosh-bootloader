package bosh

import (
	"bytes"
	"fmt"
	"net/http"
)

func (c Client) SetVMResurrection(deploymentName, jobName string, jobIndex int, enable bool) error {
	request, err := http.NewRequest("PUT", fmt.Sprintf("%s/deployments/%s/jobs/%s/%d/resurrection", c.config.URL, deploymentName, jobName, jobIndex), bytes.NewBuffer([]byte(fmt.Sprintf(`{"resurrection_paused": %t}`, !enable))))
	if err != nil {
		return err
	}

	request.SetBasicAuth(c.config.Username, c.config.Password)
	request.Header.Set("Content-Type", "application/json")

	resp, err := transport.RoundTrip(request)
	if err != nil {
		return err
	}

	body, err := bodyReader(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response %d %s: %s", resp.StatusCode, http.StatusText(resp.StatusCode), body)
	}

	return nil
}
