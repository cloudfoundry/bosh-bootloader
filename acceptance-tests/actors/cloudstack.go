package actors

func NewCloudStackLBHelper() vSphereLBHelper {
	return vSphereLBHelper{}
}

type cloudstackHelper struct {
}

func (cloudstackHelper) GetLBArgs() []string {
	return []string{}
}

func (cloudstackHelper) VerifyCloudConfigExtensions(vmExtensions []string) {
}

func (cloudstackHelper) ConfirmLBsExist(envID string) {
}

func (cloudstackHelper) ConfirmNoLBsExist(envID string) {
}

func (cloudstackHelper) VerifyBblLBOutput(stdout string) {
}
