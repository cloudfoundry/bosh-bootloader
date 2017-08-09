package bosh

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"golang.org/x/net/proxy"
)

type Client interface {
	UpdateCloudConfig(yaml []byte) error
	ConfigureHTTPClient(proxy.Dialer)
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

func NewClient(jumpbox bool, directorAddress, username, password, caCert string) Client {
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
		caCert:          caCert,
		jumpbox:         jumpbox,
	}
}

func (c client) ConfigureHTTPClient(socks5Client proxy.Dialer) {
	if socks5Client != nil {
		c.httpClient.Transport = &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return socks5Client.Dial(network, addr)
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
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

		specialTransportHTTP := &http.Client{
			Transport: c.httpClient.Transport,
		}

		ctx := context.Background()
		ctx = context.WithValue(ctx, oauth2.HTTPClient, specialTransportHTTP)

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
