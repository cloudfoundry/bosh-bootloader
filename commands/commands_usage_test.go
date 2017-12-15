package commands_test

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/commands"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Commands Usage", func() {
	Describe("Up", func() {
		Describe("Usage", func() {
			It("returns string describing usage", func() {
				upCmd := commands.Up{}
				usageText := upCmd.Usage()
				Expect(usageText).To(Equal(`Deploys BOSH director on an IAAS

  --iaas                     IAAS to deploy your BOSH director onto: "aws", "azure", "gcp", "vsphere"   env: $BBL_IAAS
  --name                     Name to assign to your BOSH director (optional)                            env: $BBL_ENV_NAME

  --aws-access-key-id        AWS Access Key ID              env: $BBL_AWS_ACCESS_KEY_ID
  --aws-secret-access-key    AWS Secret Access Key          env: $BBL_AWS_SECRET_ACCESS_KEY
  --aws-region               AWS Region                     env: $BBL_AWS_REGION

  --gcp-service-account-key  GCP Service Access Key to use  env: $BBL_GCP_SERVICE_ACCOUNT_KEY
  --gcp-region               GCP Region to use              env: $BBL_GCP_REGION

  --azure-subscription-id    Azure Subscription ID          env: $BBL_AZURE_SUBSCRIPTION_ID
  --azure-tenant-id          Azure Tenant ID                env: $BBL_AZURE_TENANT_ID
  --azure-client-id          Azure Client ID                env: $BBL_AZURE_CLIENT_ID
  --azure-client-secret      Azure Client Secret            env: $BBL_AZURE_CLIENT_SECRET
  --azure-region             Azure Region                   env: $BBL_AZURE_REGION

  --vsphere-vcenter-user     vSphere vCenter User           env: $BBL_VSPHERE_VCENTER_USER
  --vsphere-vcenter-password vSphere vCenter Password       env: $BBL_VSPHERE_VCENTER_PASSWORD
  --vsphere-vcenter-ip       vSphere vCenter IP             env: $BBL_VSPHERE_VCENTER_IP
  --vsphere-vcenter-dc       vSphere vCenter Datacenter     env: $BBL_VSPHERE_VCENTER_DC
  --vsphere-vcenter-cluster  vSphere vCenter Cluster        env: $BBL_VSPHERE_VCENTER_CLUSTER
  --vsphere-vcenter-rp       vSphere vCenter Resource Pool  env: $BBL_VSPHERE_VCENTER_RP
  --vsphere-network          vSphere Network                env: $BBL_VSPHERE_NETWORK
  --vsphere-vcenter-ds       vSphere vCenter Datastore      env: $BBL_VSPHERE_VCENTER_DS
  --vsphere-subnet           vSphere Subnet                 env: $BBL_VSPHERE_SUBNET

  Load Balancer options:
  --lb-type                  Load balancer(s) type: "concourse" or "cf"
  --lb-cert                  Path to SSL certificate (supported when type="cf")
  --lb-key                   Path to SSL certificate key (supported when type="cf")
  --lb-chain                 Path to SSL certificate chain (supported when iaas="aws")
  --lb-domain                Creates a DNS zone and records for the given domain (supported when type="cf")`))
			})
		})
	})

	Describe("Plan", func() {
		Describe("Usage", func() {
			It("returns string describing usage", func() {
				planCmd := commands.Plan{}
				usageText := planCmd.Usage()
				Expect(usageText).To(Equal(fmt.Sprintf(`Populates a state directory with the latest config without applying it

  --iaas                     IAAS to deploy your BOSH director onto: "aws", "azure", "gcp", "vsphere"   env: $BBL_IAAS
  --name                     Name to assign to your BOSH director (optional)                            env: $BBL_ENV_NAME
%s%s`, commands.Credentials, commands.LBUsage)))
			})
		})
	})

	Describe("Create LBs", func() {
		Describe("Usage", func() {
			It("returns string describing usage", func() {
				command := commands.CreateLBs{}
				usageText := command.Usage()
				Expect(usageText).To(Equal(fmt.Sprintf(`Attaches load balancer(s) with a certificate, key, and optional chain

  --type              Load balancer(s) type: "concourse" or "cf"
  [--cert]            Path to SSL certificate (supported when type="cf")
  [--key]             Path to SSL certificate key (supported when type="cf")
  [--chain]           Path to SSL certificate chain (supported when iaas="aws")
  [--domain]          Creates a DNS zone and records for the given domain (supported when type="cf")

  Credentials for your IaaS are required:%s`, commands.Credentials)))
			})
		})
	})

	Describe("Delete LBs", func() {
		Describe("Usage", func() {
			It("returns string describing usage", func() {
				command := commands.DeleteLBs{}
				usageText := command.Usage()
				Expect(usageText).To(Equal(fmt.Sprintf(`Deletes load balancer(s)

  [--skip-if-missing]  Skips deleting load balancer(s) if it is not attached (optional)

  Credentials for your IaaS are required:%s`, commands.Credentials)))
			})
		})
	})

	Describe("Destroy", func() {
		Describe("Usage", func() {
			It("returns string describing usage", func() {
				command := commands.Destroy{}
				usageText := command.Usage()
				Expect(usageText).To(Equal(fmt.Sprintf(`Tears down BOSH director infrastructure

  [--no-confirm]       Do not ask for confirmation (optional)
  [--skip-if-missing]  Gracefully exit if there is no state file (optional)

  Credentials for your IaaS are required:%s`, commands.Credentials)))
			})
		})
	})

	Describe("Rotate", func() {
		Describe("Usage", func() {
			It("returns string describing usage", func() {
				command := commands.Rotate{}
				usageText := command.Usage()
				Expect(usageText).To(Equal(fmt.Sprintf(`Rotates SSH key for the jumpbox user.

  Credentials for your IaaS are required:%s`, commands.Credentials)))
			})
		})
	})

	Describe("Usage", func() {
		Describe("Usage", func() {
			It("returns string describing usage", func() {
				command := commands.Usage{}
				usageText := command.Usage()
				Expect(usageText).To(Equal(`Prints helpful message for the given command`))
			})
		})
	})

	DescribeTable("command description", func(command commands.Command, expectedDescription string) {
		usageText := command.Usage()
		Expect(usageText).To(Equal(expectedDescription))
	},
		Entry("LBs", commands.LBs{}, "Prints attached load balancer(s)"),
		Entry("jumpbox-address", newStateQuery("jumpbox address"), "Prints BOSH jumpbox address"),
		Entry("director-address", newStateQuery("director address"), "Prints BOSH director address"),
		Entry("director-password", newStateQuery("director password"), "Prints BOSH director password"),
		Entry("director-username", newStateQuery("director username"), "Prints BOSH director username"),
		Entry("director-ca-cert", newStateQuery("director ca cert"), "Prints BOSH director CA certificate"),
		Entry("env-id", newStateQuery("environment id"), "Prints environment ID"),
		Entry("ssh-key", commands.SSHKey{}, "Prints SSH private key for the jumpbox."),
		Entry("director-ssh-key", commands.SSHKey{Director: true}, "Prints SSH private key for the director."),
		Entry("print-env", commands.PrintEnv{}, "Prints required BOSH environment variables"),
		Entry("latest-error", commands.LatestError{}, "Prints the output from the latest call to terraform"),
		Entry("bosh-deployment-vars", commands.BOSHDeploymentVars{}, "Prints required variables for BOSH deployment"),
		Entry("jumpbox-deployment-vars", commands.JumpboxDeploymentVars{}, "Prints required variables for jumpbox deployment"),
		Entry("version", commands.Version{}, "Prints version"),
		Entry("cloud-config", commands.CloudConfig{}, "Prints suggested cloud configuration for BOSH environment"),
	)
})

func newStateQuery(propertyName string) commands.StateQuery {
	return commands.NewStateQuery(nil, nil, nil, propertyName)
}
