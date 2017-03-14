package bosh

type CreateEnvError struct {
	boshState map[string]interface{}
	err       error
}

func NewCreateEnvError(boshState map[string]interface{}, err error) CreateEnvError {
	return CreateEnvError{
		boshState: boshState,
		err:       err,
	}
}

func (b CreateEnvError) Error() string {
	return b.err.Error()
}

func (b CreateEnvError) BOSHState() map[string]interface{} {
	return b.boshState
}
