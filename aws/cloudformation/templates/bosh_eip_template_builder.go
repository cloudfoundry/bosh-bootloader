package templates

type BOSHEIPTemplateBuilder struct{}

func NewBOSHEIPTemplateBuilder() BOSHEIPTemplateBuilder {
	return BOSHEIPTemplateBuilder{}
}

func (t BOSHEIPTemplateBuilder) BOSHEIP() Template {
	return Template{
		Resources: map[string]Resource{
			"BOSHEIP": Resource{
				Type: "AWS::EC2::EIP",
				Properties: EIP{
					Domain: "vpc",
				},
			},
		},
		Outputs: map[string]Output{
			"BOSHEIP": Output{Value: Ref{"BOSHEIP"}},
		},
	}
}
