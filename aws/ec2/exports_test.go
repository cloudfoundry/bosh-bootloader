package ec2

func NewClientWithInjectedEC2Client(ec2Client EC2Client, logger logger) Client {
	return Client{
		ec2Client: ec2Client,
		logger:    logger,
	}
}

func (c Client) GetEC2Client() EC2Client {
	return c.ec2Client
}
