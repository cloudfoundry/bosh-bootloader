package templates

import "strings"

type BOSHIAMTemplateBuilder struct{}

func NewBOSHIAMTemplateBuilder() BOSHIAMTemplateBuilder {
	return BOSHIAMTemplateBuilder{}
}

func getUserName(userName string) string {
	if strings.HasPrefix(userName, "bosh-iam-user-") {
		return userName
	}
	return ""
}

func (t BOSHIAMTemplateBuilder) BOSHIAMUser(userName string) Template {
	return Template{
		Resources: map[string]Resource{
			"BOSHUser": Resource{
				Type: "AWS::IAM::User",
				Properties: IAMUser{
					UserName: getUserName(userName),
					Policies: []IAMPolicy{
						{
							PolicyName: "aws-cpi",
							PolicyDocument: IAMPolicyDocument{
								Version: "2012-10-17",
								Statement: []IAMStatement{
									{
										Action: []string{
											"ec2:AssociateAddress",
											"ec2:AttachVolume",
											"ec2:CreateVolume",
											"ec2:DeleteSnapshot",
											"ec2:DeleteVolume",
											"ec2:DescribeAddresses",
											"ec2:DescribeImages",
											"ec2:DescribeInstances",
											"ec2:DescribeRegions",
											"ec2:DescribeSecurityGroups",
											"ec2:DescribeSnapshots",
											"ec2:DescribeSubnets",
											"ec2:DescribeVolumes",
											"ec2:DetachVolume",
											"ec2:CreateSnapshot",
											"ec2:CreateTags",
											"ec2:RunInstances",
											"ec2:TerminateInstances",
											"ec2:RegisterImage",
											"ec2:DeregisterImage",
										},
										Effect:   "Allow",
										Resource: "*",
									},
									{
										Action:   []string{"elasticloadbalancing:*"},
										Effect:   "Allow",
										Resource: "*",
									},
								},
							},
						},
					},
				},
			},
			"BOSHUserAccessKey": Resource{
				Properties: IAMAccessKey{
					UserName: Ref{"BOSHUser"},
				},
				Type: "AWS::IAM::AccessKey",
			},
		},
		Outputs: map[string]Output{
			"BOSHUserAccessKey": Output{
				Value: Ref{"BOSHUserAccessKey"},
			},
			"BOSHUserSecretAccessKey": Output{
				Value: FnGetAtt{
					[]string{
						"BOSHUserAccessKey",
						"SecretAccessKey",
					},
				},
			},
		},
	}
}
