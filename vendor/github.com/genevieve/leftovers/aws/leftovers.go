package aws

import (
	"errors"
	"fmt"
	"sync"

	awslib "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	awselb "github.com/aws/aws-sdk-go/service/elb"
	awselbv2 "github.com/aws/aws-sdk-go/service/elbv2"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
	awsrds "github.com/aws/aws-sdk-go/service/rds"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/genevieve/leftovers/aws/common"
	"github.com/genevieve/leftovers/aws/ec2"
	"github.com/genevieve/leftovers/aws/elb"
	"github.com/genevieve/leftovers/aws/elbv2"
	"github.com/genevieve/leftovers/aws/iam"
	"github.com/genevieve/leftovers/aws/rds"
	"github.com/genevieve/leftovers/aws/s3"
)

type resource interface {
	List(filter string) ([]common.Deletable, error)
	ListAll(filter string) ([]common.Deletable, error)
}

type Leftovers struct {
	logger    logger
	resources []resource
}

func NewLeftovers(logger logger, accessKeyId, secretAccessKey, region string) (Leftovers, error) {
	if accessKeyId == "" {
		return Leftovers{}, errors.New("Missing aws access key id.")
	}

	if secretAccessKey == "" {
		return Leftovers{}, errors.New("Missing secret access key.")
	}

	if region == "" {
		return Leftovers{}, errors.New("Missing region.")
	}

	config := &awslib.Config{
		Credentials: credentials.NewStaticCredentials(accessKeyId, secretAccessKey, ""),
		Region:      awslib.String(region),
	}
	sess := session.New(config)

	iamClient := awsiam.New(sess)
	ec2Client := awsec2.New(sess)
	elbClient := awselb.New(sess)
	elbv2Client := awselbv2.New(sess)
	s3Client := awss3.New(sess)
	rdsClient := awsrds.New(sess)

	rolePolicies := iam.NewRolePolicies(iamClient, logger)
	userPolicies := iam.NewUserPolicies(iamClient, logger)
	accessKeys := iam.NewAccessKeys(iamClient, logger)
	internetGateways := ec2.NewInternetGateways(ec2Client, logger)
	routeTables := ec2.NewRouteTables(ec2Client, logger)
	subnets := ec2.NewSubnets(ec2Client, logger)
	bucketManager := s3.NewBucketManager(region)

	return Leftovers{
		logger: logger,
		resources: []resource{
			iam.NewInstanceProfiles(iamClient, logger),
			iam.NewRoles(iamClient, logger, rolePolicies),
			iam.NewUsers(iamClient, logger, userPolicies, accessKeys),
			iam.NewPolicies(iamClient, logger),

			ec2.NewAddresses(ec2Client, logger),
			ec2.NewKeyPairs(ec2Client, logger),
			ec2.NewInstances(ec2Client, logger),
			ec2.NewSecurityGroups(ec2Client, logger),
			ec2.NewTags(ec2Client, logger),
			ec2.NewVolumes(ec2Client, logger),
			ec2.NewNetworkInterfaces(ec2Client, logger),
			ec2.NewVpcs(ec2Client, logger, routeTables, subnets, internetGateways),

			elb.NewLoadBalancers(elbClient, logger),
			elbv2.NewLoadBalancers(elbv2Client, logger),
			elbv2.NewTargetGroups(elbv2Client, logger),

			s3.NewBuckets(s3Client, logger, bucketManager),

			rds.NewDBInstances(rdsClient, logger),
			rds.NewDBSubnetGroups(rdsClient, logger),

			iam.NewServerCertificates(iamClient, logger),
		},
	}, nil
}

func (l Leftovers) List(filter string) {
	var all []common.Deletable
	for _, r := range l.resources {
		list, err := r.ListAll(filter)
		if err != nil {
			l.logger.Println(err.Error())
		}

		all = append(all, list...)
	}

	for _, r := range all {
		l.logger.Println(fmt.Sprintf("%s: %s", r.Type(), r.Name()))
	}
}

func (l Leftovers) Delete(filter string) error {
	deletables := [][]common.Deletable{}

	for _, r := range l.resources {
		list, err := r.List(filter)
		if err != nil {
			l.logger.Println(err.Error())
		}

		deletables = append(deletables, list)
	}

	var wg sync.WaitGroup

	for _, resources := range deletables {
		for _, r := range resources {
			wg.Add(1)

			go func(r common.Deletable) {
				defer wg.Done()

				l.logger.Println(fmt.Sprintf("Deleting %s.", r.Name()))

				err := r.Delete()
				if err != nil {
					l.logger.Println(err.Error())
				} else {
					l.logger.Println(fmt.Sprintf("SUCCESS deleting %s!", r.Name()))
				}
			}(r)
		}

		wg.Wait()
	}

	return nil
}
