package storage

type Azure struct {
	ClientID       string `json:"-"`
	ClientSecret   string `json:"-"`
	Region         string `json:"region,omitempty"`
	SubscriptionID string `json:"-"`
	TenantID       string `json:"-"`
}
