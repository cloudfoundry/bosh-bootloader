package azure

type logger interface {
	Printf(m string, a ...interface{})
	Prompt(m string) bool
	Println(m string)
	NoConfirm()
}
