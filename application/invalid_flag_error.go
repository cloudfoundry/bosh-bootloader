package application

type InvalidFlagError struct {
	rawError error
}

func NewInvalidFlagError(rawError error) InvalidFlagError {
	return InvalidFlagError{
		rawError: rawError,
	}
}

func (i InvalidFlagError) Error() string {
	return i.rawError.Error()
}
