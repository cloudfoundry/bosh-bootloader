package fakes

import "github.com/cloudfoundry/bosh-bootloader/aws"

type AWSClient struct {
	RetrieveAZsCall struct {
		Receives struct {
			Region string
		}
		Returns struct {
			AZs   []string
			Error error
		}
		CallCount int
	}
	RetrieveDNSCall struct {
		Receives struct {
			URL string
		}
		Returns struct {
			DNS aws.DNSZone
		}
		CallCount int
	}
}

func (a *AWSClient) RetrieveAZs(region string) ([]string, error) {
	a.RetrieveAZsCall.Receives.Region = region
	a.RetrieveAZsCall.CallCount++
	return a.RetrieveAZsCall.Returns.AZs, a.RetrieveAZsCall.Returns.Error
}

func (a *AWSClient) RetrieveDNS(url string) aws.DNSZone {
	a.RetrieveDNSCall.Receives.URL = url
	a.RetrieveDNSCall.CallCount++
	return a.RetrieveDNSCall.Returns.DNS
}
