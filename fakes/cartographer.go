package fakes

type Cartographer struct {
	YmlizeWithPrefixCall struct {
		CallCount int
		Receives  struct {
			Tfstate string
			Prefix  string
		}
		Returns struct {
			Yml   string
			Error error
		}
	}
}

func (c *Cartographer) YmlizeWithPrefix(tfstate, prefix string) (string, error) {
	c.YmlizeWithPrefixCall.CallCount++
	c.YmlizeWithPrefixCall.Receives.Tfstate = tfstate
	c.YmlizeWithPrefixCall.Receives.Prefix = prefix

	return c.YmlizeWithPrefixCall.Returns.Yml, c.YmlizeWithPrefixCall.Returns.Error
}
