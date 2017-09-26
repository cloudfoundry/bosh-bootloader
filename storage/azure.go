package storage

type Azure struct {
	ClientID       string `json:"clientId"`
	ClientSecret   string `json:"clientSecret"`
	Location       string `json:"location"`
	SubscriptionID string `json:"subscriptionId"`
	TenantID       string `json:"tenantId"`
}
