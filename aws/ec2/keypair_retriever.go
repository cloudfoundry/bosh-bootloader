package ec2

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type KeyPairRetriever struct {
}

func NewKeyPairRetriever() KeyPairRetriever {
	return KeyPairRetriever{}
}

type KeyPairInfo struct {
	Name        string
	Fingerprint string
}

func (KeyPairRetriever) Retrieve(session Session, name string) (KeyPairInfo, bool, error) {
	params := &ec2.DescribeKeyPairsInput{
		KeyNames: []*string{
			aws.String(name),
		},
	}

	resp, err := session.DescribeKeyPairs(params)
	if err != nil {
		if strings.Contains(err.Error(), "InvalidKeyPair.NotFound") {
			return KeyPairInfo{}, false, nil
		}
		return KeyPairInfo{}, false, err
	}

	if len(resp.KeyPairs) < 1 {
		return KeyPairInfo{}, false, errors.New("insufficient keypairs have been retrieved")
	}

	keypair := resp.KeyPairs[0]

	return KeyPairInfo{
		Fingerprint: *keypair.KeyFingerprint,
		Name:        *keypair.KeyName,
	}, true, nil
}
