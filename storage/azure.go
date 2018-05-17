package storage

type Azure struct {
	ClientID                string `json:"-"`
	ClientSecret            string `json:"-"`
	Region                  string `json:"region,omitempty"`
	SubscriptionID          string `json:"-"`
	TenantID                string `json:"-"`
	ResourceGroupName       string `json:"resource_group_name,omitempty"`
	VnetResourceGroupName   string `json:"vnet_resource_group_name,omitempty"`
	VnetName                string `json:"vnet_name,omitempty"`
	SubnetName              string `json:"subnet_name,omitempty"`
}
