package application

type InvalidCommandError struct {
	rawError error
}

func NewInvalidCommandError(rawError error) InvalidCommandError {
	return InvalidCommandError{
		rawError: rawError,
	}
}

func (i InvalidCommandError) Error() string {
	return i.rawError.Error()
}
