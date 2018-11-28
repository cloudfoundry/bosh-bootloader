package actors

func NewOpenStackLBHelper() openStackLBHelper {
	return openStackLBHelper{}
}

type openStackLBHelper struct {
}

func (o openStackLBHelper) GetLBArgs() []string {
	return []string{}
}

func (o openStackLBHelper) VerifyCloudConfigExtensions(vmExtensions []string) {
}

func (o openStackLBHelper) ConfirmLBsExist(envID string) {
}

func (o openStackLBHelper) ConfirmNoLBsExist(envID string) {
}

func (o openStackLBHelper) VerifyBblLBOutput(stdout string) {
}

func (o openStackLBHelper) ConfirmNoStemcellsExist(stemcellIDs []string) {}
