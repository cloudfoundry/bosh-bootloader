package aws

type logger interface {
	Printf(m string, a ...interface{})
	Println(m string)
	Prompt(m string) bool
	PromptWithDetails(resourceType, resourceName string) bool
	NoConfirm()
}
