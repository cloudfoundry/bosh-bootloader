package ec2

import (
	"fmt"

	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
)

type Tag struct {
	client     tagsClient
	key        *string
	value      *string
	resourceId *string
	identifier string
}

func NewTag(client tagsClient, key, value, resourceId *string) Tag {
	return Tag{
		client:     client,
		key:        key,
		value:      value,
		resourceId: resourceId,
		identifier: *value,
	}
}

func (t Tag) Delete() error {
	_, err := t.client.DeleteTags(&awsec2.DeleteTagsInput{
		Tags:      []*awsec2.Tag{{Key: t.key, Value: t.value}},
		Resources: []*string{t.resourceId},
	})

	if err != nil {
		return fmt.Errorf("FAILED deleting tag %s: %s", t.identifier, err)
	}

	return nil
}
