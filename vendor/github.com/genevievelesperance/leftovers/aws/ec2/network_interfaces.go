package ec2

import (
	"fmt"
	"strings"

	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevievelesperance/leftovers/aws/common"
)

type networkInterfacesClient interface {
	DescribeNetworkInterfaces(*awsec2.DescribeNetworkInterfacesInput) (*awsec2.DescribeNetworkInterfacesOutput, error)
	DeleteNetworkInterface(*awsec2.DeleteNetworkInterfaceInput) (*awsec2.DeleteNetworkInterfaceOutput, error)
}

type NetworkInterfaces struct {
	client networkInterfacesClient
	logger logger
}

func NewNetworkInterfaces(client networkInterfacesClient, logger logger) NetworkInterfaces {
	return NetworkInterfaces{
		client: client,
		logger: logger,
	}
}

func (e NetworkInterfaces) List(filter string) ([]common.Deletable, error) {
	networkInterfaces, err := e.client.DescribeNetworkInterfaces(&awsec2.DescribeNetworkInterfacesInput{})
	if err != nil {
		return nil, fmt.Errorf("Describing network interfaces: %s", err)
	}

	var resources []common.Deletable
	for _, i := range networkInterfaces.NetworkInterfaces {
		resource := NewNetworkInterface(e.client, i.NetworkInterfaceId, i.TagSet)

		if !strings.Contains(resource.identifier, filter) {
			continue
		}

		proceed := e.logger.Prompt(fmt.Sprintf("Are you sure you want to delete network interface %s?", resource.identifier))
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
