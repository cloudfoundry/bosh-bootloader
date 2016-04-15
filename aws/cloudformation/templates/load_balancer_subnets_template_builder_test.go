package templates_test

import (
	"fmt"
	"reflect"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"

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
			template := loadBalancerSubnetsTemplateBuilder.LoadBalancerSubnets(2)

			Expect(template.Parameters).To(HaveLen(2))
			Expect(template.Parameters["LoadBalancerSubnet1CIDR"].Default).To(Equal("10.0.2.0/24"))
			Expect(template.Parameters["LoadBalancerSubnet2CIDR"].Default).To(Equal("10.0.3.0/24"))

			Expect(HasLBSubnetWithAvailabilityZoneIndex(template, 0)).To(BeTrue())
			Expect(HasLBSubnetWithAvailabilityZoneIndex(template, 1)).To(BeTrue())
		})
	})
})

func HasLBSubnetWithAvailabilityZoneIndex(template templates.Template, index int) bool {
	azIndex := fmt.Sprintf("%d", index)
	subnetName := fmt.Sprintf("LoadBalancerSubnet%d", index+1)
	subnetCIDRName := fmt.Sprintf("%sCIDR", subnetName)
	tagName := fmt.Sprintf("LoadBalancer%d", index+1)

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
