package s3

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
)

type Bucket struct {
	client     bucketsClient
	name       *string
	identifier string
}

func NewBucket(client bucketsClient, name *string) Bucket {
	return Bucket{
		client:     client,
		name:       name,
		identifier: *name,
	}
}

func (b Bucket) Delete() error {
	_, err := b.client.DeleteBucket(&awss3.DeleteBucketInput{
		Bucket: b.name,
	})

	if err != nil {
		ec2err, ok := err.(awserr.Error)

		if ok && ec2err.Code() == "BucketNotEmpty" {
			resp, err := b.client.ListObjectVersions(&awss3.ListObjectVersionsInput{Bucket: b.name})
			if err != nil {
				return err
			}

			objects := make([]*awss3.ObjectIdentifier, 0)

			if len(resp.DeleteMarkers) != 0 {
				for _, v := range resp.DeleteMarkers {
					objects = append(objects, &awss3.ObjectIdentifier{
						Key:       v.Key,
						VersionId: v.VersionId,
					})
				}
			}

			if len(resp.Versions) != 0 {
				for _, v := range resp.Versions {
					objects = append(objects, &awss3.ObjectIdentifier{
						Key:       v.Key,
						VersionId: v.VersionId,
					})
				}
			}

			_, err = b.client.DeleteObjects(&awss3.DeleteObjectsInput{
				Bucket: b.name,
				Delete: &awss3.Delete{Objects: objects},
			})
			if err != nil {
				return err
			}

			return b.Delete()
		}

		return fmt.Errorf("FAILED deleting bucket %s: %s", *b.name, err)
	}

	return nil
}
