package iam

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type accessKeysClient interface {
	ListAccessKeys(*awsiam.ListAccessKeysInput) (*awsiam.ListAccessKeysOutput, error)
	DeleteAccessKey(*awsiam.DeleteAccessKeyInput) (*awsiam.DeleteAccessKeyOutput, error)
}

type accessKeys interface {
	Delete(userName string) error
}

type AccessKeys struct {
	client accessKeysClient
	logger logger
}

func NewAccessKeys(client accessKeysClient, logger logger) AccessKeys {
	return AccessKeys{
		client: client,
		logger: logger,
	}
}

func (k AccessKeys) Delete(userName string) error {
	accessKeys, err := k.client.ListAccessKeys(&awsiam.ListAccessKeysInput{UserName: aws.String(userName)})
	if err != nil {
		return fmt.Errorf("Listing access keys: %s", err)
	}

	for _, a := range accessKeys.AccessKeyMetadata {
		n := *a.AccessKeyId

		_, err = k.client.DeleteAccessKey(&awsiam.DeleteAccessKeyInput{
			UserName:    aws.String(userName),
			AccessKeyId: a.AccessKeyId,
		})
		if err == nil {
			k.logger.Printf("SUCCESS deleting access key %s\n", n)
		} else {
			k.logger.Printf("ERROR deleting access key %s: %s\n", n, err)
		}
	}

	return nil
}
