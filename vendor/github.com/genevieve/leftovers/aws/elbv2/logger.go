package elbv2

type logger interface {
	PromptWithDetails(resourceType, resourceName string) bool
}
