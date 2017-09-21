package bosh

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"

	yaml "gopkg.in/yaml.v2"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	"golang.org/x/net/proxy"
)

var proxySOCKS5 func(string, string, *proxy.Auth, proxy.Dialer) (proxy.Dialer, error) = proxy.SOCKS5

type ClientProvider struct {
	socks5Proxy socks5Proxy
}

func NewClientProvider(socks5Proxy socks5Proxy) ClientProvider {
	return ClientProvider{
		socks5Proxy: socks5Proxy,
	}
}

func (c ClientProvider) Dialer(jumpbox storage.Jumpbox) (proxy.Dialer, error) {
	privateKey, err := getJumpboxSSHKey(jumpbox.Variables)
	if err != nil {
		return nil, fmt.Errorf("get jumpbox ssh key: %s", err)
	}

	err = c.socks5Proxy.Start(privateKey, jumpbox.URL)
	if err != nil {
		return nil, fmt.Errorf("start proxy: %s", err)
	}

	socks5Dialer, err := proxySOCKS5("tcp", c.socks5Proxy.Addr(), nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("create socks5 client: %s", err)
	}

	return socks5Dialer, nil
}

func (ClientProvider) HTTPClient(dialer proxy.Dialer, directorCACert []byte) *http.Client {
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(directorCACert)
	return &http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return dialer.Dial(network, addr)
			},
			TLSClientConfig: &tls.Config{
				RootCAs: pool,
			},
		},
	}
}

func (c ClientProvider) Client(jumpbox storage.Jumpbox, directorAddress, directorUsername, directorPassword, directorCACert string) (Client, error) {
	dialer, err := c.Dialer(jumpbox)
	if err != nil {
		// not tested
		return client{}, err
	}

	httpClient := c.HTTPClient(dialer, []byte(directorCACert))
	boshClient := NewClient(httpClient, directorAddress, directorUsername, directorPassword, directorCACert)
	return boshClient, nil
}

func getJumpboxSSHKey(vars string) (string, error) {
	var variables struct {
		JumpboxSSH struct {
			PrivateKey string `yaml:"private_key"`
		} `yaml:"jumpbox_ssh"`
	}

	err := yaml.Unmarshal([]byte(vars), &variables)
	if err != nil {
		return "", err
	}

	if variables.JumpboxSSH.PrivateKey == "" {
		return "", errors.New("private key not found")
	}

	return variables.JumpboxSSH.PrivateKey, nil
}
