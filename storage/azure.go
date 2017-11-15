package storage

type Azure struct {
	ClientID       string `json:"-"`
	ClientSecret   string `json:"-"`
	Location       string `json:"location"`
	SubscriptionID string `json:"-"`
	TenantID       string `json:"-"`
}
