package elb

type logger interface {
	PromptWithDetails(resourceType, resourceName string) bool
}
