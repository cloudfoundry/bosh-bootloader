package commands

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"time"

	"github.com/pivotal-cf-experimental/bosh-bootloader/commands/aws"
	"golang.org/x/crypto/ssh"
)

type UnsupportedCreateBoshAWSKeypairCommand struct {
	*AWSConfig `no-flag:"true"`
}

func (c *UnsupportedCreateBoshAWSKeypairCommand) Execute(args []string) error {
	return CreateAndUploadRSAKey(rand.Reader, aws.NewEc2Client(aws.Config{
		AccessKeyID:      c.AWSAccessKeyID,
		SecretAccessKey:  c.AWSSecretAccessKey,
		Region:           c.AWSRegion,
		EndpointOverride: c.EndpointOverride,
	}))
}

func CreateAndUploadRSAKey(random io.Reader, ec2Client aws.Ec2) error {
	rsakey, err := rsa.GenerateKey(random, 2048)
	if err != nil {
		return err
	}

	pub, err := ssh.NewPublicKey(rsakey.Public())
	if err != nil {
		return err
	}

	keyName := fmt.Sprintf("keypair-%d", time.Now().Unix())

	err = ec2Client.ImportPublicKey(keyName, ssh.MarshalAuthorizedKey(pub))
	if err != nil {
		return err
	}

	return nil
}
