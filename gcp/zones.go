package gcp

type Zones struct {
	clientProvider clientProvider
}

func NewZones(clientProvider clientProvider) Zones {
	return Zones{
		clientProvider: clientProvider,
	}
}

func (z Zones) Get(region string) ([]string, error) {
	client := z.clientProvider.Client()
	zoneList, err := client.GetZones(region)
	if err != nil {
		return []string{}, err
	}

	zones := []string{}
	for _, zone := range zoneList.Items {
		zones = append(zones, zone.Name)
	}

	return zones, nil
}
