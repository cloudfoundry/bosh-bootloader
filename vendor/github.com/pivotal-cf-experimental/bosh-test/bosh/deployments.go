package bosh

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Deployment struct {
	Name        string
	Releases    []Release
	Stemcells   []Stemcell
	CloudConfig string
}

func (c Client) Deployments() ([]Deployment, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/deployments", c.config.URL), nil)
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

		return nil, fmt.Errorf("unexpected response %d %s:\n%s", response.StatusCode, http.StatusText(response.StatusCode), body)
	}

	var jsonDeployments []struct {
		Name     string
		Releases []struct {
			Name    string
			Version string
		}
		Stemcells []struct {
			Name    string
			Version string
		}
		CloudConfig string `json:"cloud_config"`
	}

	err = json.NewDecoder(response.Body).Decode(&jsonDeployments)
	if err != nil {
		return nil, err
	}

	var deployments []Deployment
	for _, deployment := range jsonDeployments {
		releaseMap := map[string][]string{}
		for _, release := range deployment.Releases {
			releaseMap[release.Name] = append(releaseMap[release.Name], release.Version)
		}

		var releases []Release
		for name, versions := range releaseMap {
			releases = append(releases, Release{
				Name:     name,
				Versions: versions,
			})
		}

		stemcellMap := map[string][]string{}
		for _, stemcell := range deployment.Stemcells {
			stemcellMap[stemcell.Name] = append(stemcellMap[stemcell.Name], stemcell.Version)
		}

		var stemcells []Stemcell
		for name, versions := range stemcellMap {
			stemcells = append(stemcells, Stemcell{
				Name:     name,
				Versions: versions,
			})
		}

		deployments = append(deployments, Deployment{
			Name:        deployment.Name,
			Releases:    releases,
			Stemcells:   stemcells,
			CloudConfig: deployment.CloudConfig,
		})
	}
	return deployments, nil
}
