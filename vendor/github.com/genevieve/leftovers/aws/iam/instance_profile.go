package iam

import (
	"fmt"
	"strings"

	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type InstanceProfile struct {
	client     instanceProfilesClient
	name       *string
	identifier string
}

func NewInstanceProfile(client instanceProfilesClient, name *string, roles []*awsiam.Role) InstanceProfile {
	identifier := *name

	extra := []string{}
	for _, r := range roles {
		extra = append(extra, fmt.Sprintf("Role:%s", *r.RoleName))
	}

	if len(extra) > 0 {
		identifier = fmt.Sprintf("%s (%s)", *name, strings.Join(extra, ", "))
	}

	return InstanceProfile{
		client:     client,
		name:       name,
		identifier: identifier,
	}
}

func (i InstanceProfile) Delete() error {
	_, err := i.client.DeleteInstanceProfile(&awsiam.DeleteInstanceProfileInput{
		InstanceProfileName: i.name,
	})

	if err != nil {
		return fmt.Errorf("FAILED deleting instance profile %s: %s", i.identifier, err)
	}

	return nil
}

func (i InstanceProfile) Name() string {
	return i.identifier
}
