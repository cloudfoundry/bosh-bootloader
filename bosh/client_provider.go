package bosh

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	"golang.org/x/net/proxy"
)

var proxySOCKS5 func(string, string, *proxy.Auth, proxy.Dialer) (proxy.Dialer, error) = proxy.SOCKS5

type ClientProvider struct {
	socks5Proxy  socks5Proxy
	sshKeyGetter sshKeyGetter
}

type socks5Proxy interface {
	Start(username string, key string, address string) error
	Addr() (string, error)
}

func NewClientProvider(socks5Proxy socks5Proxy, sshKeyGetter sshKeyGetter) ClientProvider {
	return ClientProvider{
		socks5Proxy:  socks5Proxy,
		sshKeyGetter: sshKeyGetter,
	}
}

func (c ClientProvider) Dialer(jumpbox storage.Jumpbox) (proxy.Dialer, error) {
	privateKey, err := c.sshKeyGetter.Get("jumpbox")
	if err != nil {
		return nil, fmt.Errorf("get jumpbox ssh key: %s", err)
	}

	err = c.socks5Proxy.Start("", privateKey, jumpbox.URL)
	if err != nil {
		return nil, fmt.Errorf("start proxy: %s", err)
	}

	addr, err := c.socks5Proxy.Addr()
	if err != nil {
		return nil, fmt.Errorf("get proxy address: %s", err)
	}

	socks5Dialer, err := proxySOCKS5("tcp", addr, nil, proxy.Direct)
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
