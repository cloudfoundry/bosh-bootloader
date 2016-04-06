package bosh

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Client interface {
	UpdateCloudConfig(yaml []byte) error
	Info() (Info, error)
}

type Info struct {
	Name    string `json:"name"`
	UUID    string `json:"uuid"`
	Version string `json:"version"`
}

type client struct {
	directorAddress string
	username        string
	password        string
	httpClient      *http.Client
}

func NewClient(directorAddress, username, password string) Client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	return client{
		directorAddress: directorAddress,
		username:        username,
		password:        password,
		httpClient:      httpClient,
	}
}

func (c client) Info() (Info, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/info", c.directorAddress), strings.NewReader(""))
	if err != nil {
		return Info{}, err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return Info{}, err
	}

	if response.StatusCode != http.StatusOK {
		return Info{}, fmt.Errorf("unexpected http response %d %s", response.StatusCode, http.StatusText(response.StatusCode))
	}

	var info Info
	if err := json.NewDecoder(response.Body).Decode(&info); err != nil {
		return Info{}, err
	}

	return info, nil
}

func (c client) UpdateCloudConfig(yaml []byte) error {
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/cloud_configs", c.directorAddress), bytes.NewBuffer(yaml))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "text/yaml")
	request.SetBasicAuth(c.username, c.password)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected http response %d %s", response.StatusCode, http.StatusText(response.StatusCode))
	}

	return nil
}
