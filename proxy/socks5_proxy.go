package proxy

import (
	"net"

	socks5 "github.com/armon/go-socks5"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/context"
)

type Socks5Proxy struct {
}

func NewSocks5Proxy() Socks5Proxy {
	return Socks5Proxy{}
}

func (s Socks5Proxy) Start(key, url string) error {
	signer, err := ssh.ParsePrivateKey([]byte(key))
	if err != nil {
		panic(err)
	}

	clientConfig := &ssh.ClientConfig{
		User: "jumpbox",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		//TODO: Change this to something more secure and test it
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	serverConn, err := ssh.Dial("tcp", url, clientConfig)
	if err != nil {
		panic(err)
	}

	conf := &socks5.Config{
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			//not tested
			return serverConn.Dial(network, addr)
		},
	}
	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}

	go func() {
		err = server.ListenAndServe("tcp", "127.0.0.1:9999")
		if err != nil {
			panic(err)
		}
	}()

	return nil
}

func (s Socks5Proxy) Stop() error {
	return nil
}
