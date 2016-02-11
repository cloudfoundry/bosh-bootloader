package aws

import (
	goaws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type Ec2ClientInterface interface {
	ImportKeyPair(input *ec2.ImportKeyPairInput) (*ec2.ImportKeyPairOutput, error)
}

type Ec2 struct {
	Client Ec2ClientInterface
}

type Config struct {
	AccessKeyID      string
	SecretAccessKey  string
	Region           string
	EndpointOverride string
}

func NewEc2Client(config Config) Ec2 {
	awsConfig := &goaws.Config{
		Credentials: credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, ""),
		Region:      goaws.String(config.Region),
	}

	if config.EndpointOverride != "" {
		awsConfig.WithEndpoint(config.EndpointOverride)
	}

	return Ec2{
		Client: ec2.New(session.New(), awsConfig),
	}
}

func (e Ec2) ImportPublicKey(name string, publicKey []byte) error {
	params := &ec2.ImportKeyPairInput{
		KeyName:           goaws.String(name),
		PublicKeyMaterial: publicKey,
	}

	_, err := e.Client.ImportKeyPair(params)
	if err != nil {
		return err
	}

	return nil
}
