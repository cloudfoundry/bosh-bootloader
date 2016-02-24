package fakes

import "github.com/aws/aws-sdk-go/service/ec2"

type EC2Session struct{}

func (s *EC2Session) ImportKeyPair(*ec2.ImportKeyPairInput) (*ec2.ImportKeyPairOutput, error) {
	return nil, nil
}

func (s *EC2Session) DescribeKeyPairs(*ec2.DescribeKeyPairsInput) (*ec2.DescribeKeyPairsOutput, error) {
	return nil, nil
}
