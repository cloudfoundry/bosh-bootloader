package compute

import (
	"fmt"
	"strings"
	"sync"

	gcpcompute "google.golang.org/api/compute/v1"
)

type firewallsClient interface {
	ListFirewalls() (*gcpcompute.FirewallList, error)
	DeleteFirewall(firewall string) error
}

type Firewalls struct {
	client firewallsClient
	logger logger
}

func NewFirewalls(client firewallsClient, logger logger) Firewalls {
	return Firewalls{
		client: client,
		logger: logger,
	}
}

func (f Firewalls) List(filter string) (map[string]string, error) {
	firewalls, err := f.client.ListFirewalls()
	if err != nil {
		return nil, fmt.Errorf("Listing firewalls: %s", err)
	}

	delete := map[string]string{}
	for _, firewall := range firewalls.Items {
		if !strings.Contains(firewall.Name, filter) {
			continue
		}

		proceed := f.logger.Prompt(fmt.Sprintf("Are you sure you want to delete firewall %s?", firewall.Name))
		if !proceed {
			continue
		}

		delete[firewall.Name] = ""
	}

	return delete, nil
}

func (f Firewalls) Delete(firewalls map[string]string) {
	var wg sync.WaitGroup

	for name, _ := range firewalls {
		wg.Add(1)

		go func(name string) {
			err := f.client.DeleteFirewall(name)

			if err != nil {
				f.logger.Printf("ERROR deleting firewall %s: %s\n", name, err)
			} else {
				f.logger.Printf("SUCCESS deleting firewall %s\n", name)
			}

			wg.Done()
		}(name)
	}

	wg.Wait()
}
