package vsphere

type logger interface {
	Printf(message string, a ...interface{})
	Println(message string)
	PromptWithDetails(resourceType, resourceName string) bool
}
