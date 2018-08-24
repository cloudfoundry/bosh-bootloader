package ec2

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/aws/common"
)

type Instance struct {
	client       instancesClient
	logger       logger
	resourceTags resourceTags
	id           *string
	identifier   string
	rtype        string
}

func NewInstance(client instancesClient, logger logger, resourceTags resourceTags, id, keyName *string, tags []*awsec2.Tag) Instance {
	identifier := *id

	extra := []string{}
	for _, t := range tags {
		extra = append(extra, fmt.Sprintf("%s:%s", *t.Key, *t.Value))
	}

	if keyName != nil && *keyName != "" {
		extra = append(extra, fmt.Sprintf("KeyPairName:%s", *keyName))
	}

	if len(extra) > 0 {
		identifier = fmt.Sprintf("%s (%s)", *id, strings.Join(extra, ", "))
	}

	return Instance{
		client:       client,
		logger:       logger,
		resourceTags: resourceTags,
		id:           id,
		identifier:   identifier,
		rtype:        "EC2 Instance",
	}
}

var pending = []string{"pending", "running", "shutting-down", "stopped", "stopping"}
var target = []string{"terminated"}

// Delete finds any addresses bound to the instance set for deletion,
// terminates the instance, waits for it to be terminated, deletes
// any tags that were bound to this instance, and finally releases
// the addresses.
func (i Instance) Delete() error {
	addresses, err := i.client.DescribeAddresses(&awsec2.DescribeAddressesInput{
		Filters: []*awsec2.Filter{{
			Name:   aws.String("instance-id"),
			Values: []*string{i.id},
		}},
	})
	if err != nil {
		return fmt.Errorf("Describe addresses: %s", err)
	}

	input := &awsec2.TerminateInstancesInput{InstanceIds: []*string{i.id}}

	_, err = i.client.TerminateInstances(input)
	if err != nil {
		ec2err, ok := err.(awserr.Error)
		if ok && ec2err.Code() == "InvalidInstanceID.NotFound" {
			return nil
		}

		return fmt.Errorf("Terminate: %s", err)
	}

	refresh := instanceRefresh(i.client, i.id)
	state := common.NewState(i.logger, refresh, pending, target)

	_, err = state.Wait()
	if err != nil {
		return fmt.Errorf("Waiting for deletion: %s", err)
	}

	err = i.resourceTags.Delete("instance", *i.id)
	if err != nil {
		return fmt.Errorf("Delete resource tags: %s", err)
	}

	for _, a := range addresses.Addresses {
		_, err = i.client.ReleaseAddress(&awsec2.ReleaseAddressInput{AllocationId: a.AllocationId})
		if err != nil {
			return fmt.Errorf("Release address: %s", err)
		}
	}

	return nil
}

func (i Instance) Name() string {
	return i.identifier
}

func (i Instance) Type() string {
	return i.rtype
}

func instanceRefresh(client instancesClient, id *string) common.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := client.DescribeInstances(&awsec2.DescribeInstancesInput{
			InstanceIds: []*string{id},
		})
		if err != nil {
			if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "InvalidInstanceID.NotFound" {
				resp = nil
			} else {
				return nil, "", err
			}
		}

		if resp == nil || len(resp.Reservations) == 0 || len(resp.Reservations[0].Instances) == 0 {
			return nil, "", nil
		}

		i := resp.Reservations[0].Instances[0]
		state := *i.State.Name

		return i, state, nil
	}
}
