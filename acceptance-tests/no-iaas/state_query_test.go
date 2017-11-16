package acceptance_test

import (
	"io/ioutil"
	"path/filepath"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("state query against a bbl 5.1.0 state file", func() {
	// Tests all the bbl read commands:
	//
	//   lbs                     Prints attached load balancer(s)
	//   bosh-deployment-vars    Prints required variables for BOSH deployment
	//   jumpbox-deployment-vars Prints required variables for jumpbox deployment
	//   cloud-config            Prints suggested cloud configuration for BOSH environment
	//   jumpbox-address         Prints BOSH jumpbox address
	//   director-address        Prints BOSH director address
	//   director-username       Prints BOSH director username
	//   director-password       Prints BOSH director password
	//   director-ca-cert        Prints BOSH director CA certificate
	//   env-id                  Prints environment ID
	//   latest-error            Prints the output from the latest call to terraform
	//   print-env               Prints BOSH friendly environment variables
	//   ssh-key                 Prints jumpbox SSH private key
	//   director-ssh-key        Prints director SSH private key
	var (
		bbl actors.BBL
	)

	BeforeEach(func() {
		stateDir, err := ioutil.TempDir("", "")
		ioutil.WriteFile(filepath.Join(stateDir, "bbl-state.json"), []byte(BBL_STATE_5_1_0), storage.StateMode)
		Expect(err).NotTo(HaveOccurred())
		bbl = actors.NewBBL(stateDir, pathToBBL, acceptance.Config{}, "no-env")
	})

	It("bbl lbs", func() {
		stdout := bbl.Lbs()
		Expect(stdout).To(Equal(`CF Router LB: 35.201.97.214
CF SSH Proxy LB: 104.196.181.208
CF TCP Router LB: 35.185.98.78
CF WebSocket LB: 104.196.197.242
CF Credhub LB: 35.196.150.246`))
	})

	It("bbl bosh-deployment-vars", func() {
		stdout := bbl.BOSHDeploymentVars()
		Expect(stdout).To(Equal(`internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.6
director_name: bosh-some-env-bbl5
zone: us-east1-b
network: some-env-bbl5-network
subnetwork: some-env-bbl5-subnet
tags:
- some-env-bbl5-bosh-director`))
	})

	It("bbl jumpbox-deployment vars", func() {
		stdout := bbl.JumpboxDeploymentVars()
		Expect(stdout).To(Equal(`internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.5
director_name: bosh-some-env-bbl5
external_ip: 35.185.60.196
zone: us-east1-b
network: some-env-bbl5-network
subnetwork: some-env-bbl5-subnet
tags:
- some-env-bbl5-bosh-open
- some-env-bbl5-jumpbox`))
	})

	It("bbl cloud-config", func() {
		stdout := bbl.CloudConfig()
		Expect(stdout).To(ContainSubstring("vm_extensions"))
		Expect(stdout).To(ContainSubstring("vm_types"))
		Expect(stdout).To(ContainSubstring("disk_types"))
		Expect(stdout).To(ContainSubstring("zone: us-east1-b"))
		Expect(stdout).To(ContainSubstring("zone: us-east1-c"))
		Expect(stdout).To(ContainSubstring("zone: us-east1-d"))
		Expect(stdout).To(ContainSubstring("network_name: some-env-bbl5-network"))
		Expect(stdout).To(ContainSubstring("subnetwork_name: some-env-bbl5-subnet"))
		Expect(stdout).To(ContainSubstring("backend_service: some-env-bbl5-router-lb"))
		Expect(stdout).To(ContainSubstring("target_pool: some-env-bbl5-cf-ws"))
		Expect(stdout).To(ContainSubstring("target_pool: some-env-bbl5-cf-ssh-proxy"))
		Expect(stdout).To(ContainSubstring("target_pool: some-env-bbl5-cf-tcp-router"))
	})

	It("bbl jumpbox-address", func() {
		stdout := bbl.JumpboxAddress()
		Expect(stdout).To(Equal("35.185.60.196"))
	})

	It("bbl director-address", func() {
		stdout := bbl.DirectorAddress()
		Expect(stdout).To(Equal("https://10.0.0.6:25555"))
	})

	It("bbl director-username", func() {
		stdout := bbl.DirectorUsername()
		Expect(stdout).To(Equal("admin"))
	})

	It("bbl director-password", func() {
		stdout := bbl.DirectorPassword()
		Expect(stdout).To(Equal("some-password"))
	})

	It("bbl director-ca-cert", func() {
		stdout := bbl.DirectorCACert()
		Expect(stdout).To(Equal("-----BEGIN CERTIFICATE-----\ndirector-ca-cert\n-----END CERTIFICATE-----"))
	})

	It("bbl env-id", func() {
		stdout := bbl.EnvID()
		Expect(stdout).To(Equal("some-env-bbl5"))
	})

	It("bbl latest-error", func() {
		stdout := bbl.LatestError()
		Expect(stdout).To(Equal("latest terraform error"))
	})

	It("bbl print-env", func() {
		stdout := bbl.PrintEnv()
		Expect(stdout).To(ContainSubstring("export BOSH_CLIENT=admin"))
		Expect(stdout).To(ContainSubstring("export BOSH_CLIENT_SECRET=some-password"))
		Expect(stdout).To(ContainSubstring("export BOSH_ENVIRONMENT=https://10.0.0.6:25555"))
		Expect(stdout).To(ContainSubstring("export BOSH_CA_CERT='-----BEGIN CERTIFICATE-----\ndirector-ca-cert\n-----END CERTIFICATE-----"))
		Expect(stdout).To(ContainSubstring("export BOSH_ALL_PROXY=socks5://localhost:"))
		Expect(stdout).To(ContainSubstring("export JUMPBOX_PRIVATE_KEY="))
		Expect(stdout).To(ContainSubstring("ssh -f -N -o StrictHostKeyChecking=no -o ServerAliveInterval=300 -D"))
		Expect(stdout).To(ContainSubstring("jumpbox@35.185.60.196 -i $JUMPBOX_PRIVATE_KEY"))
	})

	It("bbl ssh-key", func() {
		stdout := bbl.SSHKey()
		Expect(stdout).To(Equal("-----BEGIN RSA PRIVATE KEY-----\nssh-key\n-----END RSA PRIVATE KEY-----"))
	})

	It("bbl director-ssh-key", func() {
		stdout := bbl.DirectorSSHKey()
		Expect(stdout).To(Equal("-----BEGIN RSA PRIVATE KEY-----\ndirector-ssh-key\n-----END RSA PRIVATE KEY-----"))
	})
})

const BBL_STATE_5_1_0 = `
{
	"version": 12,
	"iaas": "gcp",
	"id": "20de3158-2e92-4a2c-6364-c05fa4907b85",
	"noDirector": false,
	"aws": {
		"region": ""
	},
	"azure": {
		"clientId": "",
		"clientSecret": "",
		"location": "",
		"subscriptionId": "",
		"tenantId": ""
	},
	"gcp": {
		"zone": "us-east1-b",
		"region": "us-east1",
		"zones": [
			"us-east1-b",
			"us-east1-c",
			"us-east1-d"
		]
	},
	"jumpbox": {
		"url": "35.185.60.196:22",
		"variables": "jumpbox_ssh:\n  private_key: |\n    -----BEGIN RSA PRIVATE KEY-----\n    ssh-key\n    -----END RSA PRIVATE KEY-----",
		"manifest": "some-jumpbox-manifest",
		"state": {}
	},
	"bosh": {
		"directorName": "bosh-some-env-bbl5",
		"directorUsername": "admin",
		"directorPassword": "some-password",
		"directorAddress": "https://10.0.0.6:25555",
		"directorSSLCA": "-----BEGIN CERTIFICATE-----\ndirector-ca-cert\n-----END CERTIFICATE-----\n",
		"directorSSLCertificate": "-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----\n",
		"directorSSLPrivateKey": "-----BEGIN RSA PRIVATE KEY-----\n-----END RSA PRIVATE KEY-----\n",
		"variables": "jumpbox_ssh:\n  private_key: |\n    -----BEGIN RSA PRIVATE KEY-----\n    director-ssh-key\n    -----END RSA PRIVATE KEY-----",
		"state": {},
		"manifest": "some-manifest",
		"userOpsFile": "some-ops-file"
	},
	"envID": "some-env-bbl5",
	"tfState": "{\n    \"version\": 3,\n    \"terraform_version\": \"0.10.7\",\n    \"serial\": 2,\n    \"lineage\": \"\",\n    \"modules\": [\n        {\n            \"path\": [\n                \"root\"\n            ],\n            \"outputs\": {\n                \"bosh_director_tag_name\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"some-env-bbl5-bosh-director\"\n                },\n                \"bosh_open_tag_name\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"some-env-bbl5-bosh-open\"\n                },\n                \"credhub_lb_ip\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"35.196.150.246\"\n                },\n                \"credhub_target_pool\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"some-env-bbl5-credhub\"\n                },\n                \"credhub_target_tags\": {\n                    \"sensitive\": false,\n                    \"type\": \"list\",\n                    \"value\": [\n                        \"some-env-bbl5-credhub\"\n                    ]\n                },\n                \"director_address\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"https://35.185.125.178:25555\"\n                },\n                \"external_ip\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"35.185.60.196\"\n                },\n                \"internal_tag_name\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"some-env-bbl5-internal\"\n                },\n                \"jumpbox_tag_name\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"some-env-bbl5-jumpbox\"\n                },\n                \"jumpbox_url\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"35.185.60.196:22\"\n                },\n                \"network_name\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"some-env-bbl5-network\"\n                },\n                \"router_backend_service\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"some-env-bbl5-router-lb\"\n                },\n                \"router_lb_ip\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"35.201.97.214\"\n                },\n                \"ssh_proxy_lb_ip\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"104.196.181.208\"\n                },\n                \"ssh_proxy_target_pool\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"some-env-bbl5-cf-ssh-proxy\"\n                },\n                \"subnetwork_name\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"some-env-bbl5-subnet\"\n                },\n                \"tcp_router_lb_ip\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"35.185.98.78\"\n                },\n                \"tcp_router_target_pool\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"some-env-bbl5-cf-tcp-router\"\n                },\n                \"ws_lb_ip\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"104.196.197.242\"\n                },\n                \"ws_target_pool\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"some-env-bbl5-cf-ws\"\n                }\n            },\n            \"resources\": {},\n            \"depends_on\": []\n        }\n    ]\n}\n",
	"lb": {
		"type": "cf",
		"cert": "-----BEGIN CERTIFICATE-----\nsome-lb-cert\n-----END CERTIFICATE-----\n",
		"key": "-----BEGIN RSA PRIVATE KEY-----\nsome-lb-key\n-----END RSA PRIVATE KEY-----\n",
		"chain": ""
	},
	"latestTFOutput": "latest terraform error"
}
`
