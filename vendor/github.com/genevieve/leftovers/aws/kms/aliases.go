package kms

import (
	"fmt"
	"strings"

	awskms "github.com/aws/aws-sdk-go/service/kms"
	"github.com/genevieve/leftovers/aws/common"
)

type aliasesClient interface {
	ListAliases(*awskms.ListAliasesInput) (*awskms.ListAliasesOutput, error)
	DeleteAlias(*awskms.DeleteAliasInput) (*awskms.DeleteAliasOutput, error)
}

type Aliases struct {
	client aliasesClient
	logger logger
}

func NewAliases(client aliasesClient, logger logger) Aliases {
	return Aliases{
		client: client,
		logger: logger,
	}
}

func (a Aliases) ListOnly(filter string) ([]common.Deletable, error) {
	return a.get(filter)
}

func (a Aliases) List(filter string) ([]common.Deletable, error) {
	resources, err := a.get(filter)
	if err != nil {
		return nil, err
	}

	var delete []common.Deletable
	for _, r := range resources {
		proceed := a.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		delete = append(delete, r)
	}

	return delete, nil
}

func (a Aliases) get(filter string) ([]common.Deletable, error) {
	aliases, err := a.client.ListAliases(&awskms.ListAliasesInput{})
	if err != nil {
		return nil, fmt.Errorf("Listing KMS Aliases: %s", err)
	}

	var resources []common.Deletable
	for _, alias := range aliases.Aliases {
		resource := NewAlias(a.client, alias.AliasName)

		if !strings.Contains(resource.Name(), filter) {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
