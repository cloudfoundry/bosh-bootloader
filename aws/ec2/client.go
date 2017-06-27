package ec2

import (
	"github.com/cloudfoundry/bosh-bootloader/aws"

	"github.com/aws/aws-sdk-go/aws/session"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
)

type Client interface {
	ImportKeyPair(*awsec2.ImportKeyPairInput) (*awsec2.ImportKeyPairOutput, error)
	DescribeKeyPairs(*awsec2.DescribeKeyPairsInput) (*awsec2.DescribeKeyPairsOutput, error)
	CreateKeyPair(*awsec2.CreateKeyPairInput) (*awsec2.CreateKeyPairOutput, error)
	DescribeAvailabilityZones(*awsec2.DescribeAvailabilityZonesInput) (*awsec2.DescribeAvailabilityZonesOutput, error)
	DeleteKeyPair(*awsec2.DeleteKeyPairInput) (*awsec2.DeleteKeyPairOutput, error)
	DescribeInstances(*awsec2.DescribeInstancesInput) (*awsec2.DescribeInstancesOutput, error)
	DescribeVpcs(*awsec2.DescribeVpcsInput) (*awsec2.DescribeVpcsOutput, error)
}

func NewClient(config aws.Config) Client {
	return awsec2.New(session.New(config.ClientConfig()))
}
