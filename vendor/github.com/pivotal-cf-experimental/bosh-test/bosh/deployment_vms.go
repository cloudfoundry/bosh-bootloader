package bosh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type VM struct {
	ID      string   `json:"id"`
	Index   int      `json:"index"`
	State   string   `json:"job_state"`
	JobName string   `json:"job_name"`
	IPs     []string `json:"ips"`
}

func (c Client) DeploymentVMs(name string) ([]VM, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/deployments/%s/vms?format=full", c.config.URL, name), nil)
	if err != nil {
		return []VM{}, err
	}

	request.SetBasicAuth(c.config.Username, c.config.Password)
	response, err := transport.RoundTrip(request)
	if err != nil {
		return []VM{}, err
	}

	if response.StatusCode != http.StatusFound {
		body, err := bodyReader(response.Body)
		if err != nil {
			return []VM{}, err
		}
		defer response.Body.Close()

		return []VM{}, fmt.Errorf("unexpected response %d %s:\n%s", response.StatusCode, http.StatusText(response.StatusCode), body)
	}

	location := response.Header.Get("Location")

	_, err = c.checkTaskStatus(location)
	if err != nil {
		return []VM{}, err
	}

	location, err = c.rewriteURL(location)
	if err != nil {
		return []VM{}, err
	}

	request, err = http.NewRequest("GET", fmt.Sprintf("%s/output?type=result", location), nil)
	if err != nil {
		return []VM{}, err
	}

	request.SetBasicAuth(c.config.Username, c.config.Password)
	response, err = transport.RoundTrip(request)
	if err != nil {
		return []VM{}, err
	}

	body, err := bodyReader(response.Body)
	if err != nil {
		return []VM{}, err
	}
	defer response.Body.Close()

	body = bytes.TrimSpace(body)
	parts := bytes.Split(body, []byte("\n"))

	var vms []VM
	for _, part := range parts {
		var vm VM
		err = json.Unmarshal(part, &vm)
		if err != nil {
			return vms, err
		}

		vms = append(vms, vm)
	}

	return vms, nil
}
