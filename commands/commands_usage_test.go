package commands_test

import (
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

  --iaas                     IAAS to deploy your BOSH director onto. Valid options: "aws", "azure", "gcp" (Defaults to environment variable BBL_IAAS)
  [--name]                   Name to assign to your BOSH director (optional, will be randomly generated)
  [--ops-file]               Path to BOSH ops file (optional)
  [--no-director]            Skips creating BOSH environment

  --aws-access-key-id        AWS Access Key ID to use (Defaults to environment variable BBL_AWS_ACCESS_KEY_ID)
  --aws-secret-access-key    AWS Secret Access Key to use (Defaults to environment variable BBL_AWS_SECRET_ACCESS_KEY)
  --aws-region               AWS Region to use (Defaults to environment variable BBL_AWS_REGION)

  --gcp-service-account-key  GCP Service Access Key to use (Defaults to environment variable BBL_GCP_SERVICE_ACCOUNT_KEY)
  --gcp-region               GCP Region to use (Defaults to environment variable BBL_GCP_REGION)

  --azure-subscription-id    Azure Subscription ID to use (Defaults to environment variable BBL_AZURE_SUBSCRIPTION_ID)
  --azure-tenant-id          Azure Tenant ID to use (Defaults to environment variable BBL_AZURE_TENANT_ID)
  --azure-client-id          Azure Client ID to use (Defaults to environment variable BBL_AZURE_CLIENT_ID)
  --azure-client-secret      Azure Client Secret to use (Defaults to environment variable BBL_AZURE_CLIENT_SECRET)
  --azure-region             Azure Region to use (Defaults to environment variable BBL_AZURE_REGION)`))
			})
		})
	})

	Describe("Plan", func() {
		Describe("Usage", func() {
			It("returns string describing usage", func() {
				planCmd := commands.Plan{}
				usageText := planCmd.Usage()
				Expect(usageText).To(Equal(`Populates a state directory with the latest config without applying it

  --iaas                     IAAS to deploy your BOSH director onto. Valid options: "aws", "azure", "gcp" (Defaults to environment variable BBL_IAAS)
  [--name]                   Name to assign to your BOSH director (optional, will be randomly generated)
  [--ops-file]               Path to BOSH ops file (optional)
  [--no-director]            Skips creating BOSH environment

  --aws-access-key-id        AWS Access Key ID to use (Defaults to environment variable BBL_AWS_ACCESS_KEY_ID)
  --aws-secret-access-key    AWS Secret Access Key to use (Defaults to environment variable BBL_AWS_SECRET_ACCESS_KEY)
  --aws-region               AWS Region to use (Defaults to environment variable BBL_AWS_REGION)

  --gcp-service-account-key  GCP Service Access Key to use (Defaults to environment variable BBL_GCP_SERVICE_ACCOUNT_KEY)
  --gcp-region               GCP Region to use (Defaults to environment variable BBL_GCP_REGION)

  --azure-subscription-id    Azure Subscription ID to use (Defaults to environment variable BBL_AZURE_SUBSCRIPTION_ID)
  --azure-tenant-id          Azure Tenant ID to use (Defaults to environment variable BBL_AZURE_TENANT_ID)
  --azure-client-id          Azure Client ID to use (Defaults to environment variable BBL_AZURE_CLIENT_ID)
  --azure-client-secret      Azure Client Secret to use (Defaults to environment variable BBL_AZURE_CLIENT_SECRET)
  --azure-region             Azure Location to use (Defaults to environment variable BBL_AZURE_REGION)`))
			})
		})
	})

	Describe("Create LBs", func() {
		Describe("Usage", func() {
			It("returns string describing usage", func() {
				command := commands.CreateLBs{}
				usageText := command.Usage()
				Expect(usageText).To(Equal(`Attaches load balancer(s) with a certificate, key, and optional chain

  --type              Load balancer(s) type. Valid options: "concourse" or "cf"
  [--cert]            Path to SSL certificate (conditionally required; refer to table below)
  [--key]             Path to SSL certificate key (conditionally required; refer to table below)
  [--chain]           Path to SSL certificate chain (optional; only supported on aws)
  [--domain]          Creates a DNS zone and records for the given domain (supported when type="cf")

  Credentials for your IaaS are required:
  --aws-access-key-id        AWS Access Key ID to use (Defaults to environment variable BBL_AWS_ACCESS_KEY_ID)
  --aws-secret-access-key    AWS Secret Access Key to use (Defaults to environment variable BBL_AWS_SECRET_ACCESS_KEY)

  --gcp-service-account-key  GCP Service Access Key to use (Defaults to environment variable BBL_GCP_SERVICE_ACCOUNT_KEY)

  --azure-subscription-id    Azure Subscription ID to use (Defaults to environment variable BBL_AZURE_SUBSCRIPTION_ID)
  --azure-tenant-id          Azure Tenant ID to use (Defaults to environment variable BBL_AZURE_TENANT_ID)
  --azure-client-id          Azure Client ID to use (Defaults to environment variable BBL_AZURE_CLIENT_ID)
  --azure-client-secret      Azure Client Secret to use (Defaults to environment variable BBL_AZURE_CLIENT_SECRET)

  --cert/--key are required for cf LBs and are not required or used for concourse LBs.`))
			})
		})
	})

	Describe("Delete LBs", func() {
		Describe("Usage", func() {
			It("returns string describing usage", func() {
				command := commands.DeleteLBs{}
				usageText := command.Usage()
				Expect(usageText).To(Equal(`Deletes load balancer(s)

  [--skip-if-missing]  Skips deleting load balancer(s) if it is not attached (optional)

  Credentials for your IaaS are required:
  --aws-access-key-id        AWS Access Key ID to use (Defaults to environment variable BBL_AWS_ACCESS_KEY_ID)
  --aws-secret-access-key    AWS Secret Access Key to use (Defaults to environment variable BBL_AWS_SECRET_ACCESS_KEY)

  --gcp-service-account-key  GCP Service Access Key to use (Defaults to environment variable BBL_GCP_SERVICE_ACCOUNT_KEY)

  --azure-subscription-id    Azure Subscription ID to use (Defaults to environment variable BBL_AZURE_SUBSCRIPTION_ID)
  --azure-tenant-id          Azure Tenant ID to use (Defaults to environment variable BBL_AZURE_TENANT_ID)
  --azure-client-id          Azure Client ID to use (Defaults to environment variable BBL_AZURE_CLIENT_ID)
  --azure-client-secret      Azure Client Secret to use (Defaults to environment variable BBL_AZURE_CLIENT_SECRET)`))
			})
		})
	})

	Describe("Destroy", func() {
		Describe("Usage", func() {
			It("returns string describing usage", func() {
				command := commands.Destroy{}
				usageText := command.Usage()
				Expect(usageText).To(Equal(`Tears down BOSH director infrastructure

  [--no-confirm]       Do not ask for confirmation (optional)
  [--skip-if-missing]  Gracefully exit if there is no state file (optional)

  Credentials for your IaaS are required:
  --aws-access-key-id        AWS Access Key ID to use (Defaults to environment variable BBL_AWS_ACCESS_KEY_ID)
  --aws-secret-access-key    AWS Secret Access Key to use (Defaults to environment variable BBL_AWS_SECRET_ACCESS_KEY)

  --gcp-service-account-key  GCP Service Access Key to use (Defaults to environment variable BBL_GCP_SERVICE_ACCOUNT_KEY)

  --azure-subscription-id    Azure Subscription ID to use (Defaults to environment variable BBL_AZURE_SUBSCRIPTION_ID)
  --azure-tenant-id          Azure Tenant ID to use (Defaults to environment variable BBL_AZURE_TENANT_ID)
  --azure-client-id          Azure Client ID to use (Defaults to environment variable BBL_AZURE_CLIENT_ID)
  --azure-client-secret      Azure Client Secret to use (Defaults to environment variable BBL_AZURE_CLIENT_SECRET)`))
			})
		})
	})

	Describe("Rotate", func() {
		Describe("Usage", func() {
			It("returns string describing usage", func() {
				command := commands.Rotate{}
				usageText := command.Usage()
				Expect(usageText).To(Equal(`Rotates SSH key for the jumpbox user.

  Credentials for your IaaS are required:
  --aws-access-key-id        AWS Access Key ID to use (Defaults to environment variable BBL_AWS_ACCESS_KEY_ID)
  --aws-secret-access-key    AWS Secret Access Key to use (Defaults to environment variable BBL_AWS_SECRET_ACCESS_KEY)

  --gcp-service-account-key  GCP Service Access Key to use (Defaults to environment variable BBL_GCP_SERVICE_ACCOUNT_KEY)

  --azure-subscription-id    Azure Subscription ID to use (Defaults to environment variable BBL_AZURE_SUBSCRIPTION_ID)
  --azure-tenant-id          Azure Tenant ID to use (Defaults to environment variable BBL_AZURE_TENANT_ID)
  --azure-client-id          Azure Client ID to use (Defaults to environment variable BBL_AZURE_CLIENT_ID)
  --azure-client-secret      Azure Client Secret to use (Defaults to environment variable BBL_AZURE_CLIENT_SECRET)`))
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
