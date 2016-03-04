package templates

type SSHKeyPairTemplateBuilder struct{}

func NewSSHKeyPairTemplateBuilder() SSHKeyPairTemplateBuilder {
	return SSHKeyPairTemplateBuilder{}
}

func (t SSHKeyPairTemplateBuilder) SSHKeyPairName(keyPairName string) Template {
	return Template{
		Parameters: map[string]Parameter{
			"SSHKeyPairName": Parameter{
				Type:        "AWS::EC2::KeyPair::KeyName",
				Default:     keyPairName,
				Description: "SSH KeyPair to use for instances",
			},
		},
	}
}
