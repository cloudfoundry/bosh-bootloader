package bosh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Lock struct {
	Type     string   `json:"type"`
	Resource []string `json:"resource"`
	Timeout  string   `json:"timeout"`
}

func (c Client) Locks() ([]Lock, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/locks", c.config.URL), bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, err
	}

	request.SetBasicAuth(c.config.Username, c.config.Password)

	response, err := transport.RoundTrip(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response %d %s", response.StatusCode, http.StatusText(response.StatusCode))
	}

	var locks []Lock
	err = json.NewDecoder(response.Body).Decode(&locks)
	if err != nil {
		return nil, err
	}

	return locks, nil
}
