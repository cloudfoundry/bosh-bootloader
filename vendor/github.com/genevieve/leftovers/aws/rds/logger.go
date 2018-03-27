package rds

type logger interface {
	PromptWithDetails(resourceType, resourceName string) bool
}
