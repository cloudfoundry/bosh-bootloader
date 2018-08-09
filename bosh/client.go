package bosh

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type ConfigUpdater interface {
	UpdateCloudConfig(yaml []byte) error
	Info() (Info, error)
}

type RuntimeConfigUpdater interface {
	UpdateRuntimeConfig(filepath, name string) error
}

type Info struct {
	Name    string `json:"name"`
	UUID    string `json:"uuid"`
	Version string `json:"version"`
}

var (
	MAX_RETRIES = 5
	RETRY_DELAY = 10 * time.Second
)

type Client struct {
	DirectorAddress string
	UAAAddress      string
	username        string
	password        string
	caCert          string
	httpClient      *http.Client
}

func NewClient(httpClient *http.Client, DirectorAddress, uaaAddress, username, password, caCert string) Client {
	return Client{
		DirectorAddress: DirectorAddress,
		UAAAddress:      uaaAddress,
		username:        username,
		password:        password,
		caCert:          caCert,
		httpClient:      httpClient,
	}
}

func (c Client) Info() (Info, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/info", c.DirectorAddress), strings.NewReader(""))
	if err != nil {
		return Info{}, err
	}

	response, err := makeRequests(c.httpClient, request)
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

type ConfigRequestBody struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

func (c Client) UpdateConfig(configType, name string, content []byte) error {
	var body bytes.Buffer

	config := ConfigRequestBody{
		Name:    name,
		Type:    configType,
		Content: string(content),
	}

	var err error
	err = json.NewEncoder(&body).Encode(config)
	if err != nil {
		return err // untested
	}

	request, err := http.NewRequest("POST", fmt.Sprintf("%s/configs", c.DirectorAddress), &body)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	ctx := context.Background()
	ctx = context.WithValue(ctx, oauth2.HTTPClient, c.httpClient)

	conf := &clientcredentials.Config{
		ClientID:     c.username,
		ClientSecret: c.password,
		TokenURL:     fmt.Sprintf("%s/oauth/token", c.UAAAddress),
	}

	httpClient := conf.Client(ctx)
	response, err := makeRequests(httpClient, request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected http response %d %s", response.StatusCode, http.StatusText(response.StatusCode))
	}

	return nil
}

func (c Client) UpdateCloudConfig(yaml []byte) error {
	return c.UpdateConfig("cloud", "default", yaml)
}

func makeRequests(httpClient *http.Client, request *http.Request) (*http.Response, error) {
	var (
		response *http.Response
		err      error
	)

	for i := 0; i < MAX_RETRIES; i++ {
		response, err = httpClient.Do(request)
		if err == nil {
			break
		}
		time.Sleep(RETRY_DELAY)
	}
	if err != nil {
		return &http.Response{}, fmt.Errorf("made %d attempts, last error: %s", MAX_RETRIES, err)
	}

	return response, nil
}
