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

  --iaas                     IAAS to deploy your BOSH Director onto. Valid options: "gcp", "aws" (Defaults to environment variable BBL_IAAS)
  [--name]                   Name to assign to your BOSH Director (optional, will be randomly generated)
  [--ops-file]               Path to BOSH ops file (optional)
  [--no-director]            Skips creating BOSH environment

  --aws-access-key-id        AWS Access Key ID to use (Defaults to environment variable BBL_AWS_ACCESS_KEY_ID)
  --aws-secret-access-key    AWS Secret Access Key to use (Defaults to environment variable BBL_AWS_SECRET_ACCESS_KEY)
  --aws-region               AWS region to use (Defaults to environment variable BBL_AWS_REGION)
  [--aws-bosh-az]            AWS availability zone to use for BOSH director (Defaults to environment variable BBL_AWS_BOSH_AZ)

  --gcp-service-account-key  GCP Service Access Key to use (Defaults to environment variable BBL_GCP_SERVICE_ACCOUNT_KEY)
  --gcp-project-id           GCP Project ID to use (Defaults to environment variable BBL_GCP_PROJECT_ID)
  --gcp-zone                 GCP Zone to use (Defaults to environment variable BBL_GCP_ZONE)
  --gcp-region               GCP Region to use (Defaults to environment variable BBL_GCP_REGION)`))
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
  [--chain]           Path to SSL certificate chain (optional; applicable if --cert/--key are required; refer to table below)
  [--domain]          Creates a nameserver with a zone for given domain (supported when type="cf")
  [--skip-if-exists]  Skips creating load balancer(s) if it is already attached (optional)

  --cert/--key requirements:
  ------------------------------
  |     | cf       | concourse |
  ------------------------------
  | aws | required | required  |
  ------------------------------
  | gcp | required | n/a       |
  ------------------------------`))
			})
		})
	})

	Describe("Update LBs", func() {
		Describe("Usage", func() {
			It("returns string describing usage", func() {
				command := commands.UpdateLBs{}
				usageText := command.Usage()
				Expect(usageText).To(Equal(`Updates load balancer(s) with the supplied certificate, key, and optional chain

  --cert               Path to SSL certificate
  --key                Path to SSL certificate key
  [--chain]            Path to SSL certificate chain (optional)
  [--domain]           Updates domain in the nameserver zone (supported when type="cf", optional)
  [--skip-if-missing]  Skips updating load balancer(s) if it is not attached (optional)`))
			})
		})
	})

	Describe("Delete LBs", func() {
		Describe("Usage", func() {
			It("returns string describing usage", func() {
				command := commands.DeleteLBs{}
				usageText := command.Usage()
				Expect(usageText).To(Equal(`Deletes load balancer(s)

  [--skip-if-missing]  Skips deleting load balancer(s) if it is not attached (optional)`))
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
  [--skip-if-missing]  Gracefully exit if there is no state file (optional)`))
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
		Entry("director-address", newStateQuery("director address"), "Prints BOSH director address"),
		Entry("director-password", newStateQuery("director password"), "Prints BOSH director password"),
		Entry("director-username", newStateQuery("director username"), "Prints BOSH director username"),
		Entry("director-ca-cert", newStateQuery("director ca cert"), "Prints BOSH director CA certificate"),
		Entry("env-id", newStateQuery("environment id"), "Prints environment ID"),
		Entry("ssh-key", commands.SSHKey{}, "Prints SSH private key for the jumpbox user. This can be used to ssh to the director/use the director as a gateway host."),
		Entry("print-env", commands.PrintEnv{}, "Prints required BOSH environment variables"),
		Entry("latest-error", commands.LatestError{}, "Prints the output from the latest call to terraform"),
		Entry("bosh-deployment-vars", commands.BOSHDeploymentVars{}, "Prints required variables for BOSH deployment"),
		Entry("version", commands.Version{}, "Prints version"),
		Entry("cloud-config", commands.CloudConfig{}, "Prints suggested cloud configuration for BOSH environment"),
	)
})

func newStateQuery(propertyName string) commands.StateQuery {
	return commands.NewStateQuery(nil, nil, nil, nil, propertyName)
}
