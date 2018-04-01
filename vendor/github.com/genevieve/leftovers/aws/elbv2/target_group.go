package elbv2

import (
	"fmt"

	awselbv2 "github.com/aws/aws-sdk-go/service/elbv2"
)

type TargetGroup struct {
	client     targetGroupsClient
	name       *string
	arn        *string
	identifier string
	rtype      string
}

func NewTargetGroup(client targetGroupsClient, name, arn *string) TargetGroup {
	return TargetGroup{
		client:     client,
		name:       name,
		arn:        arn,
		identifier: *name,
		rtype:      "ELBV2 Target Group",
	}
}

func (t TargetGroup) Delete() error {
	_, err := t.client.DeleteTargetGroup(&awselbv2.DeleteTargetGroupInput{
		TargetGroupArn: t.arn,
	})
	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (t TargetGroup) Name() string {
	return t.identifier
}

func (t TargetGroup) Type() string {
	return t.rtype
}
