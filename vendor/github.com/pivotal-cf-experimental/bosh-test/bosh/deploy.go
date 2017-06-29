package bosh

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
)

func (c Client) Deploy(manifest []byte) (int, error) {
	if len(manifest) == 0 {
		return 0, errors.New("a valid manifest is required to deploy")
	}

	request, err := http.NewRequest("POST", fmt.Sprintf("%s/deployments", c.config.URL), bytes.NewBuffer(manifest))
	if err != nil {
		return 0, err
	}

	request.Header.Set("Content-Type", "text/yaml")
	request.SetBasicAuth(c.config.Username, c.config.Password)

	response, err := transport.RoundTrip(request)
	if err != nil {
		return 0, err
	}

	body, err := bodyReader(response.Body)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusFound {
		return 0, fmt.Errorf("unexpected response %d %s:\n%s", response.StatusCode, http.StatusText(response.StatusCode), body)
	}

	return c.checkTaskStatus(response.Header.Get("Location"))
}
