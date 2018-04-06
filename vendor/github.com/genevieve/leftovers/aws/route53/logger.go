package route53

type logger interface {
	PromptWithDetails(resourceType, resourceName string) bool
}
