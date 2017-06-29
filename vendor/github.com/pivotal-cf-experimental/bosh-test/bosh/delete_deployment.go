package bosh

import (
	"errors"
	"fmt"
	"net/http"
)

func (c Client) DeleteDeployment(name string) error {
	if name == "" {
		return errors.New("a valid deployment name is required")
	}

	request, err := http.NewRequest("DELETE", fmt.Sprintf("%s/deployments/%s?force=true", c.config.URL, name), nil)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "text/yaml")
	request.SetBasicAuth(c.config.Username, c.config.Password)

	response, err := transport.RoundTrip(request)
	if err != nil {
		return err
	}

	body, err := bodyReader(response.Body)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusFound {
		return fmt.Errorf("unexpected response %d %s:\n%s", response.StatusCode, http.StatusText(response.StatusCode), body)
	}

	_, err = c.checkTaskStatus(response.Header.Get("Location"))
	return err
}
