package templates

import "fmt"

type InternalSubnetsTemplateBuilder struct{}

func NewInternalSubnetsTemplateBuilder() InternalSubnetsTemplateBuilder {
	return InternalSubnetsTemplateBuilder{}
}

func (InternalSubnetsTemplateBuilder) InternalSubnets(numberOfAvailabilityZones int) Template {
	internalSubnetTemplateBuilder := NewInternalSubnetTemplateBuilder()

	template := Template{}
	for index := 1; index <= numberOfAvailabilityZones; index++ {
		template = template.Merge(internalSubnetTemplateBuilder.InternalSubnet(
			index-1,
			fmt.Sprintf("%d", index),
			fmt.Sprintf("10.0.%d.0/20", 16*(index)),
		))
	}

	return template
}
