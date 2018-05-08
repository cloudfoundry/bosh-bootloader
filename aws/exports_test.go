package aws

func NewClientWithInjectedEC2Client(ec2Client EC2Client, logger logger) Client {
	return Client{
		ec2Client: ec2Client,
		logger:    logger,
	}
}

func NewClientWithInjectedRoute53Client(route53Client Route53Client, logger logger) Client {
	return Client{
		route53Client: route53Client,
		logger:        logger,
	}
}

func (c Client) GetEC2Client() EC2Client {
	return c.ec2Client
}

func (c Client) GetRoute53Client() Route53Client {
	return c.route53Client
}
