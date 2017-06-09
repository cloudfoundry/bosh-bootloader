package bosh

type BOSHVersionError struct {
	err error
}

func NewBOSHVersionError(err error) BOSHVersionError {
	return BOSHVersionError{
		err: err,
	}
}

func (b BOSHVersionError) Error() string {
	return b.err.Error()
}
