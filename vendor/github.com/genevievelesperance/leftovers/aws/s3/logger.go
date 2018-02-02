package s3

type logger interface {
	Printf(m string, a ...interface{})
	Prompt(m string) bool
}
