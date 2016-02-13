package commands

type Command interface {
	Execute(GlobalFlags) error
}
