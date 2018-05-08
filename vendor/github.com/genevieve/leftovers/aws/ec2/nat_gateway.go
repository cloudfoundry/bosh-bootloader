package ec2

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/aws/common"
)

type NatGateway struct {
	client     natGatewaysClient
	logger     logger
	id         *string
	identifier string
	official   string
}

func NewNatGateway(client natGatewaysClient, logger logger, id *string, tags []*awsec2.Tag) NatGateway {
	identifier := *id

	var extra []string
	for _, t := range tags {
		extra = append(extra, fmt.Sprintf("%s:%s", *t.Key, *t.Value))
	}

	if len(extra) > 0 {
		identifier = fmt.Sprintf("%s (%s)", *id, strings.Join(extra, ", "))
	}

	return NatGateway{
		client:     client,
		logger:     logger,
		id:         id,
		identifier: identifier,
		official:   "EC2 Nat Gateway",
	}
}

func (n NatGateway) Delete() error {
	_, err := n.client.DeleteNatGateway(&awsec2.DeleteNatGatewayInput{NatGatewayId: n.id})
	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	refresh := natGatewayRefresh(n.client, n.id)

	state := common.NewState(n.logger, refresh, []string{"deleting"}, []string{"deleted"})

	_, err = state.Wait()
	if err != nil {
		return fmt.Errorf("Waiting for deletion: %s", err)
	}

	return nil
}

func (n NatGateway) Name() string {
	return n.identifier
}

func (n NatGateway) Type() string {
	return n.official
}

func natGatewayRefresh(client natGatewaysClient, id *string) common.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &awsec2.DescribeNatGatewaysInput{NatGatewayIds: []*string{id}}

		resp, err := client.DescribeNatGateways(input)
		if err != nil {
			if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "NatGatewayNotFound" {
				return nil, "", nil
			} else {
				return nil, "", err
			}
		}

		ng := resp.NatGateways[0]
		return ng, *ng.State, nil
	}
}
