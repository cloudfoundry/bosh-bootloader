package aws

import (
	"errors"
	"fmt"

	awslib "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	awseks "github.com/aws/aws-sdk-go/service/eks"
	awselb "github.com/aws/aws-sdk-go/service/elb"
	awselbv2 "github.com/aws/aws-sdk-go/service/elbv2"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
	awskms "github.com/aws/aws-sdk-go/service/kms"
	awsrds "github.com/aws/aws-sdk-go/service/rds"
	awsroute53 "github.com/aws/aws-sdk-go/service/route53"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	awssts "github.com/aws/aws-sdk-go/service/sts"
	"github.com/fatih/color"
	"github.com/genevieve/leftovers/app"
	"github.com/genevieve/leftovers/aws/ec2"
	"github.com/genevieve/leftovers/aws/eks"
	"github.com/genevieve/leftovers/aws/elb"
	"github.com/genevieve/leftovers/aws/elbv2"
	"github.com/genevieve/leftovers/aws/iam"
	"github.com/genevieve/leftovers/aws/kms"
	"github.com/genevieve/leftovers/aws/rds"
	"github.com/genevieve/leftovers/aws/route53"
	"github.com/genevieve/leftovers/aws/s3"
	"github.com/genevieve/leftovers/common"
)

type resource interface {
	List(filter string) ([]common.Deletable, error)
	Type() string
}

type Leftovers struct {
	asyncDeleter app.AsyncDeleter
	logger       logger
	resources    []resource
}

// NewLeftovers returns a new Leftovers for AWS that can be used to list resources,
// list types, or delete resources for the provided account. It returns an error
// if the credentials provided are invalid.
func NewLeftovers(logger logger, accessKeyId, secretAccessKey, sessionToken, region string) (Leftovers, error) {
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
		Credentials: credentials.NewStaticCredentials(accessKeyId, secretAccessKey, sessionToken),
		Region:      awslib.String(region),
	}
	sess := session.New(config)

	eksClient := awseks.New(sess)
	ec2Client := awsec2.New(sess)
	elbClient := awselb.New(sess)
	elbv2Client := awselbv2.New(sess)
	kmsClient := awskms.New(sess)
	iamClient := awsiam.New(sess)
	rdsClient := awsrds.New(sess)
	route53Client := awsroute53.New(sess)
	s3Client := awss3.New(sess)
	stsClient := awssts.New(sess)

	rolePolicies := iam.NewRolePolicies(iamClient, logger)
	userPolicies := iam.NewUserPolicies(iamClient, logger)
	accessKeys := iam.NewAccessKeys(iamClient, logger)

	internetGateways := ec2.NewInternetGateways(ec2Client, logger)
	resourceTags := ec2.NewResourceTags(ec2Client)
	routeTables := ec2.NewRouteTables(ec2Client, logger, resourceTags)
	subnets := ec2.NewSubnets(ec2Client, logger, resourceTags)
	bucketManager := s3.NewBucketManager(region)

	recordSets := route53.NewRecordSets(route53Client)

	asyncDeleter := app.NewAsyncDeleter(logger)

	return Leftovers{
		logger:       logger,
		asyncDeleter: asyncDeleter,
		resources: []resource{
			elb.NewLoadBalancers(elbClient, logger),
			elbv2.NewLoadBalancers(elbv2Client, logger),
			elbv2.NewTargetGroups(elbv2Client, logger),

			iam.NewInstanceProfiles(iamClient, logger),
			iam.NewRoles(iamClient, logger, rolePolicies),
			iam.NewUsers(iamClient, logger, userPolicies, accessKeys),
			iam.NewPolicies(iamClient, logger),
			iam.NewServerCertificates(iamClient, logger),

			eks.NewClusters(eksClient, logger),

			ec2.NewKeyPairs(ec2Client, logger),
			ec2.NewInstances(ec2Client, logger, resourceTags),
			ec2.NewSecurityGroups(ec2Client, logger, resourceTags),
			ec2.NewTags(ec2Client, logger),
			ec2.NewVolumes(ec2Client, logger),
			ec2.NewNetworkInterfaces(ec2Client, logger),
			ec2.NewNatGateways(ec2Client, logger),
			ec2.NewVpcs(ec2Client, logger, routeTables, subnets, internetGateways, resourceTags),
			ec2.NewImages(ec2Client, stsClient, logger, resourceTags),
			ec2.NewAddresses(ec2Client, logger),
			ec2.NewSnapshots(ec2Client, stsClient, logger),

			s3.NewBuckets(s3Client, logger, bucketManager),

			rds.NewDBInstances(rdsClient, logger),
			rds.NewDBSubnetGroups(rdsClient, logger),
			rds.NewDBClusters(rdsClient, logger),

			kms.NewAliases(kmsClient, logger),
			kms.NewKeys(kmsClient, logger),

			route53.NewHostedZones(route53Client, logger, recordSets),
			route53.NewHealthChecks(route53Client, logger),
		},
	}, nil
}

// Types will print all the resource types that can
// be deleted on this IaaS.
func (l Leftovers) Types() {
	l.logger.NoConfirm()

	for _, r := range l.resources {
		l.logger.Println(r.Type())
	}
}

// List will print all the resources that contain
// the provided filter in the resource's identifier.
func (l Leftovers) List(filter string) {
	l.logger.NoConfirm()

	var all []common.Deletable
	for _, r := range l.resources {
		list, err := r.List(filter)
		if err != nil {
			l.logger.Println(err.Error())
		}

		all = append(all, list...)
	}

	for _, r := range all {
		l.logger.Println(fmt.Sprintf("[%s: %s]", r.Type(), r.Name()))
	}
}

// Delete will collect all resources that contain
// the provided filter in the resource's identifier, prompt
// you to confirm deletion (if enabled), and delete those
// that are selected.
func (l Leftovers) Delete(filter string) error {
	deletables := [][]common.Deletable{}

	for _, r := range l.resources {
		list, err := r.List(filter)
		if err != nil {
			l.logger.Println(color.YellowString(err.Error()))
		}

		deletables = append(deletables, list)
	}

	return l.asyncDeleter.Run(deletables)
}

// DeleteType will collect all resources of the provied type that contain
// the provided filter in the resource's identifier, prompt
// you to confirm deletion (if enabled), and delete those
// that are selected.
func (l Leftovers) DeleteType(filter, rType string) error {
	deletables := [][]common.Deletable{}

	for _, r := range l.resources {
		if r.Type() == rType {
			list, err := r.List(filter)
			if err != nil {
				l.logger.Println(color.YellowString(err.Error()))
			}

			deletables = append(deletables, list)
		}
	}

	return l.asyncDeleter.Run(deletables)
}
