package ec2

import (
	"fmt"
	"strings"

	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
)

type Vpc struct {
	client       vpcsClient
	routes       routeTables
	subnets      subnets
	gateways     internetGateways
	resourceTags resourceTags
	id           *string
	identifier   string
	rtype        string
}

func NewVpc(client vpcsClient,
	routes routeTables,
	subnets subnets,
	gateways internetGateways,
	resourceTags resourceTags,
	id *string,
	tags []*awsec2.Tag) Vpc {

	identifier := *id

	var extra []string
	for _, t := range tags {
		extra = append(extra, fmt.Sprintf("%s:%s", *t.Key, *t.Value))
	}

	if len(extra) > 0 {
		identifier = fmt.Sprintf("%s (%s)", *id, strings.Join(extra, ","))
	}

	return Vpc{
		client:       client,
		routes:       routes,
		subnets:      subnets,
		gateways:     gateways,
		resourceTags: resourceTags,
		id:           id,
		identifier:   identifier,
		rtype:        "EC2 VPC",
	}
}

func (v Vpc) Delete() error {
	err := v.routes.Delete(*v.id)
	if err != nil {
		return fmt.Errorf("Delete routes: %s", err)
	}

	err = v.subnets.Delete(*v.id)
	if err != nil {
		return fmt.Errorf("Delete subnets: %s", err)
	}

	err = v.gateways.Delete(*v.id)
	if err != nil {
		return fmt.Errorf("Delete internet gateways: %s", err)
	}

	_, err = v.client.DeleteVpc(&awsec2.DeleteVpcInput{VpcId: v.id})
	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	err = v.resourceTags.Delete("vpc", *v.id)
	if err != nil {
		return fmt.Errorf("Delete resource tags: %s", err)
	}

	return nil
}

func (v Vpc) Name() string {
	return v.identifier
}

func (v Vpc) Type() string {
	return v.rtype
}
