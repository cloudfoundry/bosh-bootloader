package templates

type MicroEIPTemplateBuilder struct{}

func NewMicroEIPTemplateBuilder() MicroEIPTemplateBuilder {
	return MicroEIPTemplateBuilder{}
}

func (t MicroEIPTemplateBuilder) MicroEIP() Template {
	return Template{
		Resources: map[string]Resource{
			"MicroEIP": Resource{
				Type: "AWS::EC2::EIP",
				Properties: EIP{
					Domain: "vpc",
				},
			},
		},
	}
}
