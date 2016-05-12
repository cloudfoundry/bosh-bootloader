package bosh

type ClientProvider struct{}

func NewClientProvider() ClientProvider {
	return ClientProvider{}
}

func (ClientProvider) Client(directorAddress, directorUsername, directorPassword string) Client {
	return NewClient(directorAddress, directorUsername, directorPassword)
}
