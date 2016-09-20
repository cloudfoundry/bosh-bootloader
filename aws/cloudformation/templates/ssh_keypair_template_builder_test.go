package templates_test

import (
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation/templates"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SSHKeyPairTemplateBuilder", func() {
	var builder templates.SSHKeyPairTemplateBuilder

	BeforeEach(func() {
		builder = templates.NewSSHKeyPairTemplateBuilder()
	})

	Describe("SSHKeyPairName", func() {
		It("returns a template containing the ssh keypair name", func() {
			ssh_keypair_name := builder.SSHKeyPairName("some-key-pair-name")

			Expect(ssh_keypair_name.Parameters).To(HaveLen(1))
			Expect(ssh_keypair_name.Parameters).To(HaveKeyWithValue("SSHKeyPairName", templates.Parameter{
				Type:        "AWS::EC2::KeyPair::KeyName",
				Default:     "some-key-pair-name",
				Description: "SSH KeyPair to use for instances",
			}))
		})
	})
})
