package bosh

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
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
	jumpbox         bool
	directorAddress string
	username        string
	password        string
	caCert          string
	httpClient      *http.Client
}

func NewClient(httpClient *http.Client, credhub bool, directorAddress, username, password, caCert string) Client {
	return client{
		directorAddress: directorAddress,
		username:        username,
		password:        password,
		caCert:          caCert,
		jumpbox:         credhub,
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

	var response *http.Response
	if c.jumpbox {
		urlParts, err := url.Parse(c.directorAddress)
		if err != nil {
			return err //not tested
		}

		boshHost, _, err := net.SplitHostPort(urlParts.Host)
		if err != nil {
			return err //not tested
		}

		ctx := context.Background()
		ctx = context.WithValue(ctx, oauth2.HTTPClient, c.httpClient)

		conf := &clientcredentials.Config{
			ClientID:     c.username,
			ClientSecret: c.password,
			TokenURL:     fmt.Sprintf("https://%s:8443/oauth/token", boshHost),
		}

		httpClient := conf.Client(ctx)

		response, err = httpClient.Do(request)
		if err != nil {
			return err
		}
	} else {
		request.SetBasicAuth(c.username, c.password)

		var err error
		response, err = c.httpClient.Do(request)
		if err != nil {
			return err
		}

	}

	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected http response %d %s", response.StatusCode, http.StatusText(response.StatusCode))
	}

	return nil
}
