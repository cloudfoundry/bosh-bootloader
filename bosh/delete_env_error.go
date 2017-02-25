package bosh

type DeleteEnvError struct {
	boshState map[string]interface{}
	err       error
}

func NewDeleteEnvError(boshState map[string]interface{}, err error) DeleteEnvError {
	return DeleteEnvError{
		boshState: boshState,
		err:       err,
	}
}

func (b DeleteEnvError) Error() string {
	return b.err.Error()
}

func (b DeleteEnvError) BOSHState() map[string]interface{} {
	return b.boshState
}
