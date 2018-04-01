package kms

import (
	"fmt"

	awskms "github.com/aws/aws-sdk-go/service/kms"
)

type Alias struct {
	client     aliasesClient
	name       *string
	identifier string
	rtype      string
}

func NewAlias(client aliasesClient, name *string) Alias {
	return Alias{
		client:     client,
		name:       name,
		identifier: *name,
		rtype:      "KMS Alias",
	}
}

func (a Alias) Delete() error {
	_, err := a.client.DeleteAlias(&awskms.DeleteAliasInput{AliasName: a.name})
	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (a Alias) Name() string {
	return a.identifier
}

func (a Alias) Type() string {
	return a.rtype
}
