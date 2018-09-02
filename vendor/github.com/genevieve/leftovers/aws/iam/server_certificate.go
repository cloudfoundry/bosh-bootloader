package iam

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type ServerCertificate struct {
	client     serverCertificatesClient
	name       *string
	identifier string
	rtype      string
}

func NewServerCertificate(client serverCertificatesClient, name *string) ServerCertificate {
	return ServerCertificate{
		client:     client,
		name:       name,
		identifier: *name,
		rtype:      "IAM Server Certificate",
	}
}

func (s ServerCertificate) Delete() error {
	return retry(10, time.Second, func() error {
		_, err := s.client.DeleteServerCertificate(&awsiam.DeleteServerCertificateInput{
			ServerCertificateName: s.name})

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "NoSuchEntity" {
					return nil
				}

				if awsErr.Code() == "DeleteConflict" {
					return awsErr
				}
			}
			return nonRetryableError{fmt.Errorf("Delete: %s", err)}
		}

		return nil
	})
}

func (s ServerCertificate) Name() string {
	return s.identifier
}

func (s ServerCertificate) Type() string {
	return s.rtype
}

func retry(attempts int, sleep time.Duration, f func() error) error {
	if err := f(); err != nil {
		if s, ok := err.(nonRetryableError); ok {
			return s.error
		}

		if attempts--; attempts > 0 {
			time.Sleep(2 * time.Second)
			return retry(attempts, 2*sleep, f)
		}

		return err
	}

	return nil
}

type nonRetryableError struct {
	error
}
