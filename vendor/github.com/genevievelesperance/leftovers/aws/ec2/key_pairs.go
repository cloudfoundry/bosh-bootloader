package ec2

import (
	"fmt"
	"strings"

	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevievelesperance/leftovers/aws/common"
)

type keyPairsClient interface {
	DescribeKeyPairs(*awsec2.DescribeKeyPairsInput) (*awsec2.DescribeKeyPairsOutput, error)
	DeleteKeyPair(*awsec2.DeleteKeyPairInput) (*awsec2.DeleteKeyPairOutput, error)
}

type KeyPairs struct {
	client keyPairsClient
	logger logger
}

func NewKeyPairs(client keyPairsClient, logger logger) KeyPairs {
	return KeyPairs{
		client: client,
		logger: logger,
	}
}

func (k KeyPairs) List(filter string) ([]common.Deletable, error) {
	keyPairs, err := k.client.DescribeKeyPairs(&awsec2.DescribeKeyPairsInput{})
	if err != nil {
		return nil, fmt.Errorf("Describing key pairs: %s", err)
	}

	var resources []common.Deletable
	for _, key := range keyPairs.KeyPairs {
		resource := NewKeyPair(k.client, key.KeyName)

		if !strings.Contains(resource.identifier, filter) {
			continue
		}

		proceed := k.logger.Prompt(fmt.Sprintf("Are you sure you want to delete key pair %s?", resource.identifier))
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
