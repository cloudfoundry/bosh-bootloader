package templates_test

import (
	"fmt"
	"reflect"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InternalSubnetsTemplateBuilder", func() {
	var internalSubnetsTemplateBuilder templates.InternalSubnetsTemplateBuilder

	BeforeEach(func() {
		internalSubnetsTemplateBuilder = templates.NewInternalSubnetsTemplateBuilder()
	})

	Describe("InternalSubnets", func() {
		It("creates internal subnets for each availability zone", func() {
			template := internalSubnetsTemplateBuilder.InternalSubnets(2)

			Expect(template.Parameters).To(HaveLen(2))
			Expect(template.Parameters["InternalSubnet1CIDR"].Default).To(Equal("10.0.16.0/20"))
			Expect(template.Parameters["InternalSubnet2CIDR"].Default).To(Equal("10.0.32.0/20"))

			Expect(HasSubnetWithAvailabilityZoneIndex(template, 0)).To(BeTrue())
			Expect(HasSubnetWithAvailabilityZoneIndex(template, 1)).To(BeTrue())
		})
	})
})

func HasSubnetWithAvailabilityZoneIndex(template templates.Template, index int) bool {
	azIndex := fmt.Sprintf("%d", index)
	subnetName := fmt.Sprintf("InternalSubnet%d", index+1)
	subnetCIDRName := fmt.Sprintf("%sCIDR", subnetName)
	tagName := fmt.Sprintf("Internal%d", index+1)

	return reflect.DeepEqual(template.Resources[subnetName].Properties, templates.Subnet{
		AvailabilityZone: map[string]interface{}{
			"Fn::Select": []interface{}{
				azIndex,
				map[string]templates.Ref{
					"Fn::GetAZs": templates.Ref{"AWS::Region"},
				},
			},
		},
		CidrBlock: templates.Ref{subnetCIDRName},
		VpcId:     templates.Ref{"VPC"},
		Tags: []templates.Tag{
			{
				Key:   "Name",
				Value: tagName,
			},
		},
	})
}
