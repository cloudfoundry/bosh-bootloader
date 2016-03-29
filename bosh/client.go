package bosh

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
)

type Client interface {
	UpdateCloudConfig(yaml []byte) error
}

type client struct {
	directorIPAddress string
	username          string
	password          string
}

func NewClient(directorIPAddress, username, password string) Client {
	return client{
		directorIPAddress: directorIPAddress,
		username:          username,
		password:          password,
	}
}

func (c client) UpdateCloudConfig(yaml []byte) error {
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/cloud_configs", c.directorIPAddress), bytes.NewBuffer(yaml))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "text/yaml")
	request.SetBasicAuth(c.username, c.password)

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected http response %d %s", response.StatusCode, http.StatusText(response.StatusCode))
	}

	return nil
}
