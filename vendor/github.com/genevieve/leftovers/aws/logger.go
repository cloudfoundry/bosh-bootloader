package aws

type logger interface {
	Printf(m string, a ...interface{})
	Println(m string)
	PromptWithDetails(resourceType, resourceName string) bool
	NoConfirm()
}
