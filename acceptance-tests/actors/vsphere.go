package actors

func NewVSphereLBHelper() vSphereLBHelper {
	return vSphereLBHelper{}
}

type vSphereLBHelper struct {
}

func (v vSphereLBHelper) GetLBArgs() []string {
	return []string{}
}

func (z vSphereLBHelper) VerifyCloudConfigExtensions(vmExtensions []string) {
}

func (v vSphereLBHelper) ConfirmLBsExist(envID string) {
}

func (v vSphereLBHelper) ConfirmNoLBsExist(envID string) {
}

func (v vSphereLBHelper) VerifyBblLBOutput(stdout string) {
}
