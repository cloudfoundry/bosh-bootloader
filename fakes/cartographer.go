package fakes

type Cartographer struct {
	YmlizeCall struct {
		CallCount int
		Receives  struct {
			Tfstate string
		}
		Returns struct {
			Yml   string
			Error error
		}
	}

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

	GetMapCall struct {
		CallCount int
		Receives  struct {
			Tfstate string
		}
		Returns struct {
			Map   map[string]interface{}
			Error error
		}
	}
}

func (c *Cartographer) Ymlize(tfstate string) (string, error) {
	c.YmlizeCall.CallCount++
	c.YmlizeCall.Receives.Tfstate = tfstate

	return c.YmlizeCall.Returns.Yml, c.YmlizeCall.Returns.Error
}

func (c *Cartographer) YmlizeWithPrefix(tfstate, prefix string) (string, error) {
	c.YmlizeWithPrefixCall.CallCount++
	c.YmlizeWithPrefixCall.Receives.Tfstate = tfstate
	c.YmlizeWithPrefixCall.Receives.Prefix = prefix

	return c.YmlizeWithPrefixCall.Returns.Yml, c.YmlizeWithPrefixCall.Returns.Error
}

func (c *Cartographer) GetMap(tfstate string) (map[string]interface{}, error) {
	c.GetMapCall.CallCount++
	c.GetMapCall.Receives.Tfstate = tfstate

	return c.GetMapCall.Returns.Map, c.GetMapCall.Returns.Error
}
