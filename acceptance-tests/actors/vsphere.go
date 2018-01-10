package actors

func NewVSphereLBHelper() vSphereLbHelper {
	return vSphereLbHelper{}
}

type vSphereLbHelper struct {
}

func (v vSphereLbHelper) GetLBArgs() []string {
	return []string{}
}

func (v vSphereLbHelper) ConfirmLBsExist(envID string) {
}

func (v vSphereLbHelper) ConfirmNoLBsExist(envID string) {
}

func (v vSphereLbHelper) VerifyBblLBOutput(stdout string) {
}
