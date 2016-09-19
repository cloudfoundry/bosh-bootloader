package commands_test

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"

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
				Expect(usageText).To(Equal(`Deploys a BOSH Director on AWS

--aws-access-key-id        AWS Access Key ID to use
--aws-secret-access-key    AWS Secret Access Key to use
--aws-region               AWS region to use`))
			})
		})
	})

	Describe("Create LBs", func() {
		Describe("Usage", func() {
			It("returns string describing usage", func() {
				command := commands.CreateLBs{}
				usageText := command.Usage()
				Expect(usageText).To(Equal(`Attaches a load balancer with a certificate, key, and optional chain

--type            Load balancer type. Valid options: "concourse" or "cf"
--cert            Path to SSL certificate
--key             Path to SSL certificate key
--chain           Path to SSL certificate chain (optional)
--skip-if-exists  Skips creating load balancer if it is already attached (optional)`))
			})
		})
	})

	Describe("Update LBs", func() {
		Describe("Usage", func() {
			It("returns string describing usage", func() {
				command := commands.UpdateLBs{}
				usageText := command.Usage()
				Expect(usageText).To(Equal(`Updates a load balancer with the supplied certificate, key, and optional chain

--cert             Path to SSL certificate
--key              Path to SSL certificate key
--chain            Path to SSL certificate chain (optional)
--skip-if-missing  Skips updating load balancer if it is not attached (optional)`))
			})
		})
	})

	Describe("Delete LBs", func() {
		Describe("Usage", func() {
			It("returns string describing usage", func() {
				command := commands.DeleteLBs{}
				usageText := command.Usage()
				Expect(usageText).To(Equal(`Deletes the load balancers

--skip-if-missing  Skips deleting load balancer if it is not attached (optional)`))
			})
		})
	})

	Describe("Destroy", func() {
		Describe("Usage", func() {
			It("returns string describing usage", func() {
				command := commands.Destroy{}
				usageText := command.Usage()
				Expect(usageText).To(Equal(`Tears down a BOSH Director environment on AWS

--no-confirm  Do not ask for confirmation (optional)`))
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
		Entry("LBs", commands.LBs{}, "Lists attached load balancers"),
		Entry("director-address", newStateQuery("director address"), "Prints the BOSH director address"),
		Entry("director-password", newStateQuery("director password"), "Prints the BOSH director password"),
		Entry("director-username", newStateQuery("director username"), "Prints the BOSH director username"),
		Entry("bosh-ca-cert", newStateQuery("bosh ca cert"), "Prints the BOSH director CA certificate"),
		Entry("env-id", newStateQuery("environment id"), "Prints the environment ID"),
		Entry("ssh-key", newStateQuery("ssh key"), "Prints the SSH private key"),
		Entry("version", commands.Version{}, "Prints version"),
	)
})

func newStateQuery(propertyName string) commands.StateQuery {
	return commands.NewStateQuery(nil, propertyName, nil)
}
