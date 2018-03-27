package ec2

import (
	"fmt"
	"strings"

	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/aws/common"
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

func (k KeyPairs) ListAll(filter string) ([]common.Deletable, error) {
	return k.get(filter)
}

func (k KeyPairs) List(filter string) ([]common.Deletable, error) {
	resources, err := k.get(filter)
	if err != nil {
		return nil, err
	}

	var delete []common.Deletable
	for _, r := range resources {
		proceed := k.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		delete = append(delete, r)
	}

	return delete, nil
}

func (k KeyPairs) get(filter string) ([]common.Deletable, error) {
	keyPairs, err := k.client.DescribeKeyPairs(&awsec2.DescribeKeyPairsInput{})
	if err != nil {
		return nil, fmt.Errorf("Describing key pairs: %s", err)
	}

	var resources []common.Deletable
	for _, key := range keyPairs.KeyPairs {
		resource := NewKeyPair(k.client, key.KeyName)

		if !strings.Contains(resource.Name(), filter) {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
