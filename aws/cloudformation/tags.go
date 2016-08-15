package cloudformation

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

type Tags []Tag

type Tag struct {
	Key   string
	Value string
}

func (t Tags) toAWSTags() []*cloudformation.Tag {
	awsTags := []*cloudformation.Tag{}
	for _, tag := range t {
		awsTags = append(awsTags, &cloudformation.Tag{
			Key:   aws.String(tag.Key),
			Value: aws.String(tag.Value),
		})
	}
	return awsTags
}
