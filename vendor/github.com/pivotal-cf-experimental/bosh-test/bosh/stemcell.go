package bosh

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/blang/semver"
)

type Stemcell struct {
	Name     string
	OS       string
	Versions []string
}

type stemcell struct {
	Name             string
	Operating_system string
	Version          string
}

func NewStemcell() Stemcell {
	return Stemcell{}
}

func (c Client) getStemcells(name string) ([]stemcell, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/stemcells", c.config.URL), nil)
	if err != nil {
		return []stemcell{}, err
	}

	request.SetBasicAuth(c.config.Username, c.config.Password)
	response, err := client.Do(request)
	if err != nil {
		return []stemcell{}, err
	}

	if response.StatusCode == http.StatusNotFound {
		return []stemcell{}, fmt.Errorf("stemcell %s could not be found", name)
	}

	if response.StatusCode != http.StatusOK {
		body, err := bodyReader(response.Body)
		if err != nil {
			return []stemcell{}, err
		}
		defer response.Body.Close()

		return []stemcell{}, fmt.Errorf("unexpected response %d %s:\n%s", response.StatusCode, http.StatusText(response.StatusCode), body)
	}

	var stemcells []stemcell

	err = json.NewDecoder(response.Body).Decode(&stemcells)
	if err != nil {
		return []stemcell{}, err
	}

	return stemcells, nil
}

func (c Client) StemcellByName(name string) (Stemcell, error) {
	stemcells, err := c.getStemcells(name)
	if err != nil {
		return Stemcell{}, err
	}

	stemcell := NewStemcell()
	stemcell.Name = name

	for _, s := range stemcells {
		if s.Name == name {
			stemcell.Versions = append(stemcell.Versions, s.Version)
		}
	}

	return stemcell, nil
}

func (c Client) StemcellByOS(os string) (Stemcell, error) {
	stemcells, err := c.getStemcells(os)
	if err != nil {
		return Stemcell{}, err
	}

	stemcell := NewStemcell()
	stemcell.OS = os

	for _, s := range stemcells {
		if s.Operating_system == os {
			stemcell.Versions = append(stemcell.Versions, s.Version)
		}
	}

	return stemcell, nil
}

func (s Stemcell) Latest() (string, error) {
	latestVersion := "0"

	if len(s.Versions) == 0 {
		return "", errors.New("no stemcell versions found, cannot get latest")
	}

	for _, version := range s.Versions {

		semVersion, err := semver.ParseTolerant(version)
		if err != nil {
			return "", err
		}

		semLatestVersion, err := semver.ParseTolerant(latestVersion)
		if err != nil {
			// Not tested
			return "", err
		}

		if semVersion.GT(semLatestVersion) {
			latestVersion = version
		}
	}

	return latestVersion, nil
}
