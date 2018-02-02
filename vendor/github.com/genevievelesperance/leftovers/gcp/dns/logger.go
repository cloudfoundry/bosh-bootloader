package dns

type logger interface {
	Printf(message string, a ...interface{})
	Println(message string)
	Prompt(message string) bool
}
