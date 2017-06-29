package bosh

import (
	"fmt"
	"io"
	"net/http"
)

func (c Client) Resource(resourceId string) (io.ReadCloser, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/resources/%s", c.config.URL, resourceId), nil)
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

		return nil, fmt.Errorf("unexpected response %s:\n%s", response.Status, body)
	}

	return response.Body, nil
}
