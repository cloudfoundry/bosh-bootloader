package templates_test

import (
	"fmt"
	"reflect"

	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation/templates"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LoadBalancerSubnetsTemplateBuilder", func() {
	var loadBalancerSubnetsTemplateBuilder templates.LoadBalancerSubnetsTemplateBuilder

	BeforeEach(func() {
		loadBalancerSubnetsTemplateBuilder = templates.NewLoadBalancerSubnetsTemplateBuilder()
	})

	Describe("LoadBalancerSubnets", func() {
		It("creates load balancer subnets for each availability zone", func() {
			template := loadBalancerSubnetsTemplateBuilder.LoadBalancerSubnets([]string{
				"some-zone-1",
				"some-zone-2",
			})

			Expect(template.Parameters).To(HaveLen(2))
			Expect(template.Parameters["LoadBalancerSubnet1CIDR"].Default).To(Equal("10.0.2.0/24"))
			Expect(template.Parameters["LoadBalancerSubnet2CIDR"].Default).To(Equal("10.0.3.0/24"))

			Expect(hasLBSubnetWithAvailabilityZoneIndex(template, 0)).To(BeTrue())
			Expect(hasLBSubnetWithAvailabilityZoneIndex(template, 1)).To(BeTrue())
		})
	})
})

func hasLBSubnetWithAvailabilityZoneIndex(template templates.Template, index int) bool {
	az := fmt.Sprintf("some-zone-%d", index+1)
	subnetName := fmt.Sprintf("LoadBalancerSubnet%d", index+1)
	subnetCIDRName := fmt.Sprintf("%sCIDR", subnetName)
	tagName := fmt.Sprintf("LoadBalancer%d", index+1)

	return reflect.DeepEqual(template.Resources[subnetName].Properties, templates.Subnet{
		AvailabilityZone: az,
		CidrBlock:        templates.Ref{subnetCIDRName},
		VpcId:            templates.Ref{"VPC"},
		Tags: []templates.Tag{
			{
				Key:   "Name",
				Value: tagName,
			},
		},
	})
}
