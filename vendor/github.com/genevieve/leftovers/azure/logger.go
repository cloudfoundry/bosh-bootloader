package azure

type logger interface {
	Printf(message string, args ...interface{})
	PromptWithDetails(resourceType, resourceName string) bool
	Println(message string)
	NoConfirm()
}
