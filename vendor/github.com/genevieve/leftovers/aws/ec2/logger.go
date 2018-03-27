package ec2

type logger interface {
	Printf(m string, a ...interface{})
	PromptWithDetails(resourceType, resourceName string) bool
}
