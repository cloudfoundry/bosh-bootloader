package templates

import "fmt"

type InternalSubnetsTemplateBuilder struct{}

func NewInternalSubnetsTemplateBuilder() InternalSubnetsTemplateBuilder {
	return InternalSubnetsTemplateBuilder{}
}

func (InternalSubnetsTemplateBuilder) InternalSubnets(availabilityZones []string) Template {
	internalSubnetTemplateBuilder := NewInternalSubnetTemplateBuilder()

	template := Template{}
	for index, az := range availabilityZones {
		template = template.Merge(internalSubnetTemplateBuilder.InternalSubnet(
			az,
			fmt.Sprintf("%d", index+1),
			fmt.Sprintf("10.0.%d.0/20", 16*(index+1)),
		))
	}

	return template
}
