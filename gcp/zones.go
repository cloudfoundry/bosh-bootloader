package gcp

type Zones struct{}

func NewZones() Zones {
	return Zones{}
}

func (Zones) Get(region string) []string {
	azs := map[string][]string{
		"us-west1":        []string{"us-west1-a", "us-west1-b"},
		"us-central1":     []string{"us-central1-a", "us-central1-b", "us-central1-c", "us-central1-f"},
		"us-east1":        []string{"us-east1-b", "us-east1-c", "us-east1-d"},
		"europe-west1":    []string{"europe-west1-b", "europe-west1-c", "europe-west1-d"},
		"asia-east1":      []string{"asia-east1-a", "asia-east1-b", "asia-east1-c"},
		"asia-northeast1": []string{"asia-northeast1-a", "asia-northeast1-b", "asia-northeast1-c"},
	}

	return azs[region]
}
