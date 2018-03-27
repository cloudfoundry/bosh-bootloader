package ec2

import (
	"fmt"
	"strings"

	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/aws/common"
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

func (e NetworkInterfaces) ListAll(filter string) ([]common.Deletable, error) {
	return e.get(filter)
}

func (e NetworkInterfaces) List(filter string) ([]common.Deletable, error) {
	resources, err := e.get(filter)
	if err != nil {
		return nil, err
	}

	var delete []common.Deletable
	for _, r := range resources {
		proceed := e.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		delete = append(delete, r)
	}

	return delete, nil
}

func (e NetworkInterfaces) get(filter string) ([]common.Deletable, error) {
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

		resources = append(resources, resource)
	}

	return resources, nil
}
