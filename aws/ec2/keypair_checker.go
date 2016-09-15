package ec2

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type KeyPairChecker struct {
	ec2ClientProvider ec2ClientProvider
}

func NewKeyPairChecker(ec2ClientProvider ec2ClientProvider) KeyPairChecker {
	return KeyPairChecker{
		ec2ClientProvider: ec2ClientProvider,
	}
}

type KeyPairInfo struct {
	Name        string
	Fingerprint string
}

func (k KeyPairChecker) HasKeyPair(name string) (bool, error) {
	if name == "" {
		return false, nil
	}

	params := &ec2.DescribeKeyPairsInput{
		KeyNames: []*string{
			aws.String(name),
		},
	}

	_, err := k.ec2ClientProvider.GetEC2Client().DescribeKeyPairs(params)
	if err != nil {
		if strings.Contains(err.Error(), "InvalidKeyPair.NotFound") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
