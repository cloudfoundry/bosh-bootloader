package common

type logger interface {
	Printf(m string, a ...interface{})
}
