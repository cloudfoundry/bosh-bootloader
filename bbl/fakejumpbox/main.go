package fakejumpbox

import (
	"fmt"
	"io"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

type JumpboxServer struct {
	listener net.Listener
}

func NewJumpboxServer() *JumpboxServer {
	jumpboxServer := &JumpboxServer{}
	var err error
	jumpboxServer.listener, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to create listener: %s", err.Error()))
	}

	return jumpboxServer
}

func (j *JumpboxServer) Start(jumpboxPrivateKey, fakeBOSHServerAddr string) {
	config := createServerConfig(jumpboxPrivateKey)

	go func() {
		for {
			nConn, err := j.listener.Accept()
			if err != nil {
				log.Fatal(fmt.Sprintf("failed to accept: %s", err.Error()))
			}
			defer nConn.Close()

			_, chans, reqs, err := ssh.NewServerConn(nConn, config)
			if err != nil {
				log.Fatal("failed to handshake: ", err)
			}
			go ssh.DiscardRequests(reqs)

			for newChannel := range chans {
				if newChannel.ChannelType() != "direct-tcpip" {
					newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
					continue
				}
				channel, _, err := newChannel.Accept()
				if err != nil {
					log.Fatalf("Could not accept channel: %v", err)
				}
				defer channel.Close()

				httpConn, err := net.Dial("tcp", fakeBOSHServerAddr)
				if err != nil {
					log.Fatalf("Could not open connection to http server: %v", err)
				}
				defer httpConn.Close()

				go io.Copy(httpConn, channel)
				go io.Copy(channel, httpConn)
			}
		}
	}()
}

func createServerConfig(sshPrivateKey string) *ssh.ServerConfig {
	signer, err := ssh.ParsePrivateKey([]byte(sshPrivateKey))
	if err != nil {
		log.Fatal("Failed to parse private key: ", err)
	}

	config := &ssh.ServerConfig{
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			if string(signer.PublicKey().Marshal()) == string(pubKey.Marshal()) {
				return nil, nil
			}
			return nil, fmt.Errorf("unknown public key for %q", c.User())
		},
	}

	config.AddHostKey(signer)

	return config
}

func (j *JumpboxServer) Addr() string {
	return j.listener.Addr().String()
}
