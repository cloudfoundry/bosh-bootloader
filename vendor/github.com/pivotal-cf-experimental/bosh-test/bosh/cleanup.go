package bosh

import (
	"fmt"
	"net/http"
	"strings"
)

func (c Client) Cleanup() (int, error) {
	body := strings.NewReader(`{"config": {"remove_all": true}}`)
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/cleanup", c.config.URL), body)
	if err != nil {
		return 0, err
	}

	request.SetBasicAuth(c.config.Username, c.config.Password)
	request.Header.Set("Content-Type", "application/json")

	response, err := transport.RoundTrip(request)
	if err != nil {
		return 0, err
	}

	if response.StatusCode != http.StatusFound {
		return 0, fmt.Errorf("unexpected response %s", response.Status)
	}

	return c.checkTaskStatus(response.Header.Get("Location"))
}
