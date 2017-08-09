package bosh

type ClientProvider struct{}

func NewClientProvider() ClientProvider {
	return ClientProvider{}
}

func (ClientProvider) Client(jumpbox bool, directorAddress, directorUsername, directorPassword, directorCACert string) Client {
	return NewClient(jumpbox, directorAddress, directorUsername, directorPassword, directorCACert)
}
