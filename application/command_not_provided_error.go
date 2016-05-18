package application

type CommandNotProvidedError struct {
	rawError error
}

func NewCommandNotProvidedError() CommandNotProvidedError {
	return CommandNotProvidedError{}
}

func (i CommandNotProvidedError) Error() string {
	return "unknown command: [EMPTY]"
}
