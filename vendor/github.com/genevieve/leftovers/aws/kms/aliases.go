package kms

import (
	"fmt"
	"strings"

	awskms "github.com/aws/aws-sdk-go/service/kms"
	"github.com/genevieve/leftovers/common"
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

func (a Aliases) List(filter string) ([]common.Deletable, error) {
	aliases, err := a.client.ListAliases(&awskms.ListAliasesInput{})
	if err != nil {
		return nil, fmt.Errorf("Listing KMS Aliases: %s", err)
	}

	var resources []common.Deletable
	for _, alias := range aliases.Aliases {
		r := NewAlias(a.client, alias.AliasName)

		if !strings.Contains(r.Name(), filter) {
			continue
		}

		proceed := a.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func (a Aliases) Type() string {
	return "kms-alias"
}
