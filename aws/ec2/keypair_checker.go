package ec2

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type KeyPairChecker struct {
}

func NewKeyPairChecker() KeyPairChecker {
	return KeyPairChecker{}
}

type KeyPairInfo struct {
	Name        string
	Fingerprint string
}

func (KeyPairChecker) HasKeyPair(client Client, name string) (bool, error) {
	if name == "" {
		return false, nil
	}

	params := &ec2.DescribeKeyPairsInput{
		KeyNames: []*string{
			aws.String(name),
		},
	}

	_, err := client.DescribeKeyPairs(params)
	if err != nil {
		if strings.Contains(err.Error(), "InvalidKeyPair.NotFound") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
