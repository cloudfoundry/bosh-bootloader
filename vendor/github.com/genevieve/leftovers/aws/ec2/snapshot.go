package ec2

import (
	"fmt"

	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
)

type Snapshot struct {
	client     snapshotsClient
	id         *string
	identifier string
}

func NewSnapshot(client snapshotsClient, id *string) Snapshot {
	return Snapshot{
		client:     client,
		id:         id,
		identifier: *id,
	}
}

func (s Snapshot) Delete() error {
	_, err := s.client.DeleteSnapshot(&awsec2.DeleteSnapshotInput{SnapshotId: s.id})
	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (s Snapshot) Name() string {
	return s.identifier
}

func (s Snapshot) Type() string {
	return "EC2 Snapshot"
}
