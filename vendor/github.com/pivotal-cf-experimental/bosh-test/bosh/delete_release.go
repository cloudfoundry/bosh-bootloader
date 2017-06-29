package bosh

import (
	"fmt"
	"net/http"
	"net/url"
)

func (c Client) DeleteRelease(name, version string) error {
	query := url.Values{}
	query.Add("version", version)

	request, err := http.NewRequest("DELETE", fmt.Sprintf("%s/releases/%s?%s", c.config.URL, name, query.Encode()), nil)
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
	if err != nil {
		return err
	}

	return nil
}
