package aws

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type InputGenerator struct {
	availabilityZoneRetriever ec2.AvailabilityZoneRetriever
}

const terraformNameCharLimit = 18

var jsonMarshal = json.Marshal

func NewInputGenerator(availabilityZoneRetriever ec2.AvailabilityZoneRetriever) InputGenerator {
	return InputGenerator{
		availabilityZoneRetriever: availabilityZoneRetriever,
	}
}

func (i InputGenerator) Generate(state storage.State) (map[string]string, error) {
	azs, err := i.availabilityZoneRetriever.RetrieveAvailabilityZones(state.AWS.Region)
	if err != nil {
		return map[string]string{}, err
	}
	zones, err := jsonMarshal(azs)
	if err != nil {
		return map[string]string{}, err
	}

	shortEnvID := state.EnvID
	if len(shortEnvID) > terraformNameCharLimit {
		sha1 := fmt.Sprintf("%x", sha1.Sum([]byte(state.EnvID)))
		shortEnvID = fmt.Sprintf("%s-%s", shortEnvID[:terraformNameCharLimit-8], sha1[:terraformNameCharLimit-11])
	}

	inputs := map[string]string{
		"env_id":                 state.EnvID,
		"short_env_id":           shortEnvID,
		"access_key":             state.AWS.AccessKeyID,
		"secret_key":             state.AWS.SecretAccessKey,
		"region":                 state.AWS.Region,
		"bosh_availability_zone": "",
		"availability_zones":     string(zones),
	}

	if state.LB.Type == "cf" || state.LB.Type == "concourse" {
		inputs["ssl_certificate"] = state.LB.Cert
		inputs["ssl_certificate_private_key"] = state.LB.Key
		inputs["ssl_certificate_chain"] = state.LB.Chain

		if state.LB.Domain != "" {
			inputs["system_domain"] = state.LB.Domain
		}
	}

	return inputs, nil
}
