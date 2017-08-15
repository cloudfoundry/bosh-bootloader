package stack_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/aws/iam"
	"github.com/cloudfoundry/bosh-bootloader/stack"
	"github.com/cloudfoundry/bosh-bootloader/stack/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
)

var _ = Describe("Migrate", func() {
	var (
		tf             *fakes.TF
		infrastructure *fakes.Infrastructure
		certificate    *fakes.Certificate
		userPolicy     *fakes.UserPolicy
		zone           *fakes.Zone
		keyPair        *fakes.KeyPair

		migrator stack.Migrator

		incomingState storage.State
	)

	BeforeEach(func() {
		tf = &fakes.TF{}
		infrastructure = &fakes.Infrastructure{}
		certificate = &fakes.Certificate{}
		userPolicy = &fakes.UserPolicy{}
		keyPair = &fakes.KeyPair{}
		zone = &fakes.Zone{}

		migrator = stack.NewMigrator(tf, infrastructure, certificate, userPolicy, zone, keyPair)

		zone.RetrieveReturns([]string{"some-az"}, nil)

		infrastructure.UpdateReturns(cloudformation.Stack{
			Outputs: map[string]string{
				"VPCID":                           "some-vpc",
				"VPCInternetGatewayID":            "some-vpc-gateway-internet-gateway",
				"NATEIP":                          "some-nat-eip",
				"NATInstance":                     "some-nat-instance",
				"NATSecurityGroup":                "some-nat-security-group",
				"BOSHEIP":                         "some-bosh-eip",
				"BOSHSecurityGroup":               "some-bosh-security-group",
				"BOSHSubnet":                      "some-bosh-subnet",
				"BOSHRouteTable":                  "some-bosh-route-table",
				"InternalSecurityGroup":           "some-internal-security-group",
				"InternalRouteTable":              "some-internal-route-table",
				"InternalSubnet1Name":             "some-internal-subnet-1",
				"InternalSubnet2Name":             "some-internal-subnet-2",
				"InternalSubnet3Name":             "some-internal-subnet-3",
				"InternalSubnet4Name":             "some-internal-subnet-4",
				"LoadBalancerSubnet1Name":         "some-lb-subnet-1",
				"LoadBalancerSubnet2Name":         "some-lb-subnet-2",
				"CFRouterInternalSecurityGroup":   "some-cf-router-internal-security-group",
				"CFRouterSecurityGroup":           "some-cf-router-security-group",
				"CFRouterLoadBalancer":            "some-cf-router-load-balancer",
				"CFSSHProxyInternalSecurityGroup": "some-cf-ssh-proxy-internal-security-group",
				"CFSSHProxySecurityGroup":         "some-cf-ssh-proxy-security-group",
				"CFSSHProxyLoadBalancer":          "some-cf-ssh-proxy-load-balancer",
				"ConcourseInternalSecurityGroup":  "some-concourse-internal-security-group",
				"ConcourseSecurityGroup":          "some-concourse-security-group",
				"ConcourseLoadBalancer":           "some-concourse-load-balancer",
				"LoadBalancerRouteTable":          "some-lb-route-table",
			},
		}, nil)

		incomingState = storage.State{
			KeyPair: storage.KeyPair{
				Name: "some-keypair",
			},
			Stack: storage.Stack{
				Name:   "some-stack",
				BOSHAZ: "some-bosh-az",
			},
			EnvID: "some-env-id",
		}
	})

	It("migrates infrastructure created by cloudformation to terraform", func() {
		tf.ImportReturns("some-magic-tfstate", nil)

		incomingState.AWS = storage.AWS{
			Region: "some-region",
		}

		state, err := migrator.Migrate(incomingState)
		Expect(err).NotTo(HaveOccurred())

		Expect(zone.RetrieveCallCount()).To(Equal(1))
		Expect(zone.RetrieveArgsForCall(0)).To(Equal("some-region"))

		Expect(certificate.DescribeCallCount()).To(Equal(0))

		Expect(infrastructure.UpdateCallCount()).To(Equal(1))

		keyPairName, availabilityZones, stackName, boshAZ, lbType, certificateARN, envID := infrastructure.UpdateArgsForCall(0)
		Expect(keyPairName).To(Equal("some-keypair"))
		Expect(availabilityZones).To(Equal([]string{"some-az"}))
		Expect(stackName).To(Equal("some-stack"))
		Expect(boshAZ).To(Equal("some-bosh-az"))
		Expect(lbType).To(Equal(""))
		Expect(certificateARN).To(Equal(""))
		Expect(envID).To(Equal("some-env-id"))

		Expect(keyPair.DeleteCallCount()).To(Equal(1))
		Expect(keyPair.DeleteArgsForCall(0)).To(Equal("some-keypair"))

		Expect(userPolicy.DeleteCallCount()).To(Equal(1))
		username, policyName := userPolicy.DeleteArgsForCall(0)
		Expect(username).To(Equal("bosh-iam-user-some-env-id"))
		Expect(policyName).To(Equal("aws-cpi"))

		Expect(infrastructure.DeleteCallCount()).To(Equal(1))
		Expect(infrastructure.DeleteArgsForCall(0)).To(Equal("some-stack"))

		Expect(state).To(Equal(storage.State{
			KeyPair: storage.KeyPair{
				Name: "some-keypair",
			},
			AWS: storage.AWS{
				Region: "some-region",
			},
			EnvID: "some-env-id",
			MigratedFromCloudFormation: true,
			TFState:                    "some-magic-tfstate",
		}))
	})

	It("passes the incoming state to import", func() {
		tf.ImportReturns("some-tfstate", nil)
		incomingState.AWS = storage.AWS{
			AccessKeyID:     "some-access-key-id",
			SecretAccessKey: "some-secret-access-key",
			Region:          "some-region",
		}

		_, err := migrator.Migrate(incomingState)
		Expect(err).NotTo(HaveOccurred())

		Expect(tf.ImportArgsForCall(0).TFState).To(Equal(""))
		Expect(tf.ImportArgsForCall(0).Creds).To(Equal(incomingState.AWS))

		Expect(tf.ImportArgsForCall(1).TFState).To(Equal("some-tfstate"))
	})

	It("maps each resource to terraform", func() {
		_, err := migrator.Migrate(incomingState)
		Expect(err).NotTo(HaveOccurred())

		importInputs := []terraform.ImportInput{}
		for _, importCall := range tf.Invocations()["Import"] {
			importInputs = append(importInputs, importCall[0].(terraform.ImportInput))
		}

		Expect(tf.ImportCallCount()).To(Equal(27))
		Expect(importInputs).To(ContainElement(terraform.ImportInput{TerraformAddr: "aws_vpc.vpc", AWSResourceID: "some-vpc"}))
		Expect(importInputs).To(ContainElement(terraform.ImportInput{TerraformAddr: "aws_internet_gateway.ig", AWSResourceID: "some-vpc-gateway-internet-gateway"}))
		Expect(importInputs).To(ContainElement(terraform.ImportInput{TerraformAddr: "aws_eip.nat_eip", AWSResourceID: "some-nat-eip"}))
		Expect(importInputs).To(ContainElement(terraform.ImportInput{TerraformAddr: "aws_instance.nat", AWSResourceID: "some-nat-instance"}))
		Expect(importInputs).To(ContainElement(terraform.ImportInput{TerraformAddr: "aws_security_group.nat_security_group", AWSResourceID: "some-nat-security-group"}))
		Expect(importInputs).To(ContainElement(terraform.ImportInput{TerraformAddr: "aws_eip.bosh_eip", AWSResourceID: "some-bosh-eip"}))
		Expect(importInputs).To(ContainElement(terraform.ImportInput{TerraformAddr: "aws_security_group.bosh_security_group", AWSResourceID: "some-bosh-security-group"}))
		Expect(importInputs).To(ContainElement(terraform.ImportInput{TerraformAddr: "aws_subnet.bosh_subnet", AWSResourceID: "some-bosh-subnet"}))
		Expect(importInputs).To(ContainElement(terraform.ImportInput{TerraformAddr: "aws_route_table.bosh_route_table", AWSResourceID: "some-bosh-route-table"}))
		Expect(importInputs).To(ContainElement(terraform.ImportInput{TerraformAddr: "aws_security_group.internal_security_group", AWSResourceID: "some-internal-security-group"}))
		Expect(importInputs).To(ContainElement(terraform.ImportInput{TerraformAddr: "aws_route_table.internal_route_table", AWSResourceID: "some-internal-route-table"}))
		Expect(importInputs).To(ContainElement(gstruct.MatchFields(gstruct.IgnoreExtras,
			gstruct.Fields{
				"TerraformAddr": MatchRegexp(`aws_subnet.internal_subnets\[\d\]`),
				"AWSResourceID": Equal("some-internal-subnet-1"),
			},
		)))
		Expect(importInputs).To(ContainElement(gstruct.MatchFields(gstruct.IgnoreExtras,
			gstruct.Fields{
				"TerraformAddr": MatchRegexp(`aws_subnet.internal_subnets\[\d\]`),
				"AWSResourceID": Equal("some-internal-subnet-2"),
			},
		)))
		Expect(importInputs).To(ContainElement(gstruct.MatchFields(gstruct.IgnoreExtras,
			gstruct.Fields{
				"TerraformAddr": MatchRegexp(`aws_subnet.internal_subnets\[\d\]`),
				"AWSResourceID": Equal("some-internal-subnet-3"),
			},
		)))
		Expect(importInputs).To(ContainElement(gstruct.MatchFields(gstruct.IgnoreExtras,
			gstruct.Fields{
				"TerraformAddr": MatchRegexp(`aws_subnet.internal_subnets\[\d\]`),
				"AWSResourceID": Equal("some-internal-subnet-4"),
			},
		)))
		Expect(importInputs).To(ContainElement(terraform.ImportInput{TerraformAddr: "aws_security_group.cf_router_lb_internal_security_group", AWSResourceID: "some-cf-router-internal-security-group"}))
		Expect(importInputs).To(ContainElement(terraform.ImportInput{TerraformAddr: "aws_security_group.cf_router_lb_security_group", AWSResourceID: "some-cf-router-security-group"}))
		Expect(importInputs).To(ContainElement(terraform.ImportInput{TerraformAddr: "aws_elb.cf_router_lb", AWSResourceID: "some-cf-router-load-balancer"}))
		Expect(importInputs).To(ContainElement(terraform.ImportInput{TerraformAddr: "aws_security_group.cf_ssh_lb_internal_security_group", AWSResourceID: "some-cf-ssh-proxy-internal-security-group"}))
		Expect(importInputs).To(ContainElement(terraform.ImportInput{TerraformAddr: "aws_security_group.cf_ssh_lb_security_group", AWSResourceID: "some-cf-ssh-proxy-security-group"}))
		Expect(importInputs).To(ContainElement(terraform.ImportInput{TerraformAddr: "aws_elb.cf_ssh_lb", AWSResourceID: "some-cf-ssh-proxy-load-balancer"}))
		Expect(importInputs).To(ContainElement(terraform.ImportInput{TerraformAddr: "aws_security_group.concourse_lb_internal_security_group", AWSResourceID: "some-concourse-internal-security-group"}))
		Expect(importInputs).To(ContainElement(terraform.ImportInput{TerraformAddr: "aws_security_group.concourse_lb_security_group", AWSResourceID: "some-concourse-security-group"}))
		Expect(importInputs).To(ContainElement(terraform.ImportInput{TerraformAddr: "aws_elb.concourse_lb", AWSResourceID: "some-concourse-load-balancer"}))
		Expect(importInputs).To(ContainElement(terraform.ImportInput{TerraformAddr: "aws_route_table.lb_route_table", AWSResourceID: "some-lb-route-table"}))
		Expect(importInputs).To(ContainElement(gstruct.MatchFields(gstruct.IgnoreExtras,
			gstruct.Fields{
				"TerraformAddr": MatchRegexp(`aws_subnet.lb_subnets\[\d\]`),
				"AWSResourceID": Equal("some-lb-subnet-1"),
			},
		)))
		Expect(importInputs).To(ContainElement(gstruct.MatchFields(gstruct.IgnoreExtras,
			gstruct.Fields{
				"TerraformAddr": MatchRegexp(`aws_subnet.lb_subnets\[\d\]`),
				"AWSResourceID": Equal("some-lb-subnet-2"),
			},
		)))
	})

	Context("when there is no stack", func() {
		BeforeEach(func() {
			incomingState.Stack = storage.Stack{
				Name: "",
			}
		})

		It("exits gracefully", func() {
			state, err := migrator.Migrate(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(certificate.DescribeCallCount()).To(Equal(0))
			Expect(zone.RetrieveCallCount()).To(Equal(0))
			Expect(infrastructure.UpdateCallCount()).To(Equal(0))
			Expect(tf.ImportCallCount()).To(Equal(0))
			Expect(infrastructure.DeleteCallCount()).To(Equal(0))

			Expect(state).To(Equal(incomingState))
		})
	})

	Context("when load balancer exists", func() {
		BeforeEach(func() {
			incomingState.Stack = storage.Stack{
				Name:            "some-stack",
				CertificateName: "some-certificate-name",
				LBType:          "cf",
			}

			certificate.DescribeReturns(iam.Certificate{
				ARN:  "some-dumb-arn",
				Name: "some-certificate-name",
			}, nil)
		})

		It("migrates infrastructure created by cloudformation to terraform", func() {
			returnedState, err := migrator.Migrate(incomingState)
			Expect(err).NotTo(HaveOccurred())

			importInputs := []terraform.ImportInput{}
			for _, importCall := range tf.Invocations()["Import"] {
				importInputs = append(importInputs, importCall[0].(terraform.ImportInput))
			}

			Expect(certificate.DescribeCallCount()).To(Equal(1))
			Expect(certificate.DescribeArgsForCall(0)).To(Equal("some-certificate-name"))

			_, _, _, _, lbType, certificateARN, _ := infrastructure.UpdateArgsForCall(0)
			Expect(lbType).To(Equal("cf"))
			Expect(certificateARN).To(Equal("some-dumb-arn"))
			Expect(returnedState.LB.Type).To(Equal(lbType))

			Expect(importInputs).To(ContainElement(terraform.ImportInput{
				TerraformAddr: "aws_iam_server_certificate.lb_cert",
				AWSResourceID: "some-certificate-name",
			}))
		})
	})

	Context("when an error occurs", func() {
		Context("when the availability zone cannot be retrieved", func() {
			It("returns an error", func() {
				zone.RetrieveReturns([]string{}, errors.New("wrong-zone"))

				_, err := migrator.Migrate(incomingState)
				Expect(err).To(MatchError("wrong-zone"))
			})
		})

		Context("when the certificate cannot be described", func() {
			It("returns an error", func() {
				incomingState.Stack.LBType = "concourse"
				certificate.DescribeReturns(iam.Certificate{}, errors.New("failed to describe"))

				_, err := migrator.Migrate(incomingState)
				Expect(err).To(MatchError("failed to describe"))
			})
		})

		Context("when the infrastructure cannot be updated", func() {
			It("returns an error", func() {
				infrastructure.UpdateReturns(cloudformation.Stack{}, errors.New("did not update"))

				_, err := migrator.Migrate(incomingState)
				Expect(err).To(MatchError("did not update"))
			})
		})

		Context("when terraform fails to import the stack", func() {
			It("returns an error", func() {
				tf.ImportReturns("", errors.New("no import"))

				_, err := migrator.Migrate(incomingState)
				Expect(err).To(MatchError("no import"))
			})
		})

		Context("when the user policy cannot be deleted", func() {
			It("returns an error", func() {
				userPolicy.DeleteReturns(errors.New("no"))

				_, err := migrator.Migrate(incomingState)
				Expect(err).To(MatchError("no"))
			})
		})

		Context("when the infrastructure cannot be deleted", func() {
			It("returns an error", func() {
				infrastructure.DeleteReturns(errors.New("no"))

				_, err := migrator.Migrate(incomingState)
				Expect(err).To(MatchError("no"))
			})
		})

		Context("when the key pair cannot be deleted", func() {
			It("returns an error", func() {
				keyPair.DeleteReturns(errors.New("keypair delete"))

				_, err := migrator.Migrate(incomingState)
				Expect(err).To(MatchError("keypair delete"))
			})
		})
	})
})
