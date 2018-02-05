package s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type bucketManager interface {
	IsInRegion(bucket string) bool
}

type BucketManager struct {
	region string
}

func NewBucketManager(region string) BucketManager {
	return BucketManager{
		region: region,
	}
}

func (u BucketManager) IsInRegion(bucket string) bool {
	sess := session.Must(session.NewSession())
	r, _ := s3manager.GetBucketRegion(aws.BackgroundContext(), sess, bucket, "us-west-1")
	return u.region == r
}
