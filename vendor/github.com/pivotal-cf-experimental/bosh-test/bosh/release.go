package bosh

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Release struct {
	Name     string
	Versions []string
}

func NewRelease() Release {
	return Release{}
}

func (c Client) Release(name string) (Release, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/releases/%s", c.config.URL, name), nil)
	if err != nil {
		return Release{}, err
	}

	request.SetBasicAuth(c.config.Username, c.config.Password)
	response, err := client.Do(request)
	if err != nil {
		return Release{}, err
	}

	if response.StatusCode == http.StatusNotFound {
		return Release{}, fmt.Errorf("release %s could not be found", name)
	}

	if response.StatusCode != http.StatusOK {
		body, err := bodyReader(response.Body)
		if err != nil {
			return Release{}, err
		}
		defer response.Body.Close()

		return Release{}, fmt.Errorf("unexpected response %d %s:\n%s", response.StatusCode, http.StatusText(response.StatusCode), body)
	}

	release := NewRelease()
	err = json.NewDecoder(response.Body).Decode(&release)
	if err != nil {
		return Release{}, err
	}

	release.Name = name

	return release, nil
}

func (r Release) Latest() string {
	// THIS ASSUMES THE VERSIONS ARE ALREADY SORTED
	return r.Versions[len(r.Versions)-1]
}
