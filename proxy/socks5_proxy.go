package proxy

import (
	"fmt"
	"net"
	"strconv"

	socks5 "github.com/armon/go-socks5"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/context"
)

var netListen = net.Listen

type Socks5Proxy struct {
	logger        logger
	hostKeyGetter hostKeyGetter
	port          int
	started       bool
}

type logger interface {
	Println(string)
}

type hostKeyGetter interface {
	Get(string, string) (ssh.PublicKey, error)
}

func NewSocks5Proxy(logger logger, hostKeyGetter hostKeyGetter, port int) *Socks5Proxy {
	return &Socks5Proxy{
		logger:        logger,
		hostKeyGetter: hostKeyGetter,
		port:          port,
		started:       false,
	}
}

func (s *Socks5Proxy) Start(key, url string) error {
	if s.started {
		return nil
	}

	signer, err := ssh.ParsePrivateKey([]byte(key))
	if err != nil {
		return err
	}

	hostKey, err := s.hostKeyGetter.Get(key, url)
	if err != nil {
		return err
	}

	clientConfig := &ssh.ClientConfig{
		User: "jumpbox",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.FixedHostKey(hostKey),
	}

	serverConn, err := ssh.Dial("tcp", url, clientConfig)
	if err != nil {
		return err
	}

	conf := &socks5.Config{
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return serverConn.Dial(network, addr)
		},
	}
	server, err := socks5.New(conf)
	if err != nil {
		// not tested
		return err
	}

	if s.port == 0 {
		s.port, err = openPort()
		if err != nil {
			return err
		}
	}
	go func() {
		err = server.ListenAndServe("tcp", fmt.Sprintf("127.0.0.1:%d", s.port))
		if err != nil {
			s.logger.Println(fmt.Sprintf("err: failed to start socks5 proxy: %s", err.Error()))
		}
	}()

	s.started = true
	return nil
}

func (s *Socks5Proxy) Addr() string {
	return fmt.Sprintf("127.0.0.1:%d", s.port)
}

func openPort() (int, error) {
	l, err := netListen("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	defer l.Close()
	_, port, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(port)
}
