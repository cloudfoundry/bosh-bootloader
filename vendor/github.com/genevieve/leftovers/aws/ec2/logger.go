package ec2

type logger interface {
	Printf(m string, a ...interface{})
	Prompt(m string) bool
}

type deletable interface {
	Delete() error
}
