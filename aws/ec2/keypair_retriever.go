package ec2

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type KeypairRetriever struct {
}

func NewKeypairRetriever() KeypairRetriever {
	return KeypairRetriever{}
}

type KeyPairInfo struct {
	Fingerprint string
	Name        string
}

var KeyPairNotFound error = errors.New("keypair not found")

func (KeypairRetriever) Retrieve(session Session, name string) (KeyPairInfo, error) {
	params := &ec2.DescribeKeyPairsInput{
		KeyNames: []*string{
			aws.String(name),
		},
	}

	resp, err := session.DescribeKeyPairs(params)
	if err != nil {
		if strings.Contains(err.Error(), "InvalidKeyPair.NotFound") {
			return KeyPairInfo{}, KeyPairNotFound
		}
		return KeyPairInfo{}, err
	}

	if len(resp.KeyPairs) < 1 {
		return KeyPairInfo{}, errors.New("insufficient keypairs have been retrieved")
	}

	keypair := resp.KeyPairs[0]

	return KeyPairInfo{
		Fingerprint: *keypair.KeyFingerprint,
		Name:        *keypair.KeyName,
	}, nil
}
