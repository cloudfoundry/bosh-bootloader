package kms

import (
	"fmt"
	"strings"

	awskms "github.com/aws/aws-sdk-go/service/kms"
)

type Key struct {
	client     keysClient
	name       *string
	identifier string
	rtype      string
}

func NewKey(client keysClient, id *string, metadata *awskms.KeyMetadata, tags []*awskms.Tag) Key {
	identifier := *id

	extra := []string{}
	if metadata != nil && *metadata.Description != "" {
		extra = append(extra, fmt.Sprintf("Description:%s", *metadata.Description))
	}

	for _, tag := range tags {
		extra = append(extra, fmt.Sprintf("%s:%s", *tag.TagKey, *tag.TagValue))
	}

	if len(extra) > 0 {
		identifier = fmt.Sprintf("%s (%s)", *id, strings.Join(extra, ", "))
	}

	return Key{
		client:     client,
		name:       id,
		identifier: identifier,
		rtype:      "KMS Key",
	}
}

func (k Key) Delete() error {
	_, err := k.client.DisableKey(&awskms.DisableKeyInput{KeyId: k.name})
	if err != nil {
		return fmt.Errorf("Disable: %s", err)
	}

	_, err = k.client.ScheduleKeyDeletion(&awskms.ScheduleKeyDeletionInput{KeyId: k.name})
	if err != nil {
		return fmt.Errorf("Schedule deletion: %s", err)
	}

	return nil
}

func (k Key) Name() string {
	return k.identifier
}

func (k Key) Type() string {
	return k.rtype
}
