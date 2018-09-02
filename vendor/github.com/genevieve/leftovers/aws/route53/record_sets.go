package route53

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	awsroute53 "github.com/aws/aws-sdk-go/service/route53"
)

type recordSetsClient interface {
	ListResourceRecordSets(*awsroute53.ListResourceRecordSetsInput) (*awsroute53.ListResourceRecordSetsOutput, error)
	ChangeResourceRecordSets(*awsroute53.ChangeResourceRecordSetsInput) (*awsroute53.ChangeResourceRecordSetsOutput, error)
}

type RecordSets struct {
	client recordSetsClient
}

func NewRecordSets(client recordSetsClient) RecordSets {
	return RecordSets{
		client: client,
	}
}

func (r RecordSets) Get(hostedZoneId *string) ([]*awsroute53.ResourceRecordSet, error) {
	var (
		records    []*awsroute53.ResourceRecordSet
		nextRecord *string
	)

	for isTruncated := true; isTruncated; {
		output, err := r.client.ListResourceRecordSets(&awsroute53.ListResourceRecordSetsInput{
			HostedZoneId:    hostedZoneId,
			StartRecordName: nextRecord,
		})
		if err != nil {
			return nil, fmt.Errorf("List Resource Record Sets: %s", err)
		}

		records = append(records, output.ResourceRecordSets...)

		isTruncated = *output.IsTruncated
		nextRecord = output.NextRecordName
	}

	return records, nil
}

func (r RecordSets) Delete(hostedZoneId *string, hostedZoneName string, records []*awsroute53.ResourceRecordSet) error {
	var changes []*awsroute53.Change
	for _, record := range records {
		if strings.TrimSuffix(*record.Name, ".") == strings.TrimSuffix(hostedZoneName, ".") && (*record.Type == "NS" || *record.Type == "SOA") {
			continue
		}
		changes = append(changes, &awsroute53.Change{
			Action:            aws.String("DELETE"),
			ResourceRecordSet: record,
		})
	}

	if len(changes) > 0 {
		_, err := r.client.ChangeResourceRecordSets(&awsroute53.ChangeResourceRecordSetsInput{
			HostedZoneId: hostedZoneId,
			ChangeBatch:  &awsroute53.ChangeBatch{Changes: changes},
		})
		if err != nil {
			return fmt.Errorf("Delete Resource Record Sets: %s", err)
		}
	}

	return nil
}
