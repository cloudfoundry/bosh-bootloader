package s3

type logger interface {
	PromptWithDetails(resourceType, resourceName string) bool
}
