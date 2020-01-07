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

var _ = Describe("state query against a bbl 6.10.46 state file", func() {
	var bbl actors.BBL

	BeforeEach(func() {
		stateDir, err := ioutil.TempDir("", "")
		ioutil.WriteFile(filepath.Join(stateDir, "bbl-state.json"), []byte(BBL_STATE_6_10_46), storage.StateMode)
		Expect(err).NotTo(HaveOccurred())
		bbl = actors.NewBBL(stateDir, pathToBBL, acceptance.Config{}, "no-env", false)
	})

	It("bbl lbs", func() {
		stdout := bbl.Lbs()
		Expect(stdout).To(Equal(`CF Router LB: 35.201.97.214
CF SSH Proxy LB: 104.196.181.208
CF TCP Router LB: 35.185.98.78
CF WebSocket LB: 104.196.197.242`))
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
		Expect(stdout).To(Equal("some-env-bbl6"))
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
		Expect(stdout).To(ContainSubstring("export JUMPBOX_PRIVATE_KEY="))
		Expect(stdout).To(ContainSubstring("export BOSH_ALL_PROXY=ssh+socks5://jumpbox@35.185.60.196:22?private-key="))
		Expect(stdout).To(ContainSubstring("bosh_jumpbox_private.key"))
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

const BBL_STATE_6_10_46 = `
{
	"version": 14,
	"iaas": "gcp",
	"id": "e02be31f-0a0a-402e-4d73-de1f955be098",
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
		"directorName": "bosh-some-env-bbl6",
		"directorUsername": "admin",
		"directorPassword": "some-password",
		"directorAddress": "https://10.0.0.6:25555",
		"directorSSLCA": "-----BEGIN CERTIFICATE-----\ndirector-ca-cert\n-----END CERTIFICATE-----\n",
		"directorSSLCertificate": "-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----\n",
		"directorSSLPrivateKey": "-----BEGIN RSA PRIVATE KEY-----\n-----END RSA PRIVATE KEY-----\n",
		"variables": "jumpbox_ssh:\n  private_key: |\n    -----BEGIN RSA PRIVATE KEY-----\n    director-ssh-key\n    -----END RSA PRIVATE KEY-----",
		"state": {},
		"manifest": "some-manifest",
		"userOpsFile": ""
	},
	"envID": "some-env-bbl6",
	"tfState": "{\n    \"version\": 3,\n    \"terraform_version\": \"0.10.7\",\n    \"serial\": 2,\n    \"lineage\": \"\",\n    \"modules\": [\n        {\n            \"path\": [\n                \"root\"\n            ],\n            \"outputs\": {\n                \"bosh_director_tag_name\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"some-env-bbl6-bosh-director\"\n                },\n                \"bosh_open_tag_name\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"some-env-bbl6-bosh-open\"\n                },\n                \"credhub_lb_ip\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"35.196.150.246\"\n                },\n                \"credhub_target_pool\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"some-env-bbl6-credhub\"\n                },\n                \"credhub_target_tags\": {\n                    \"sensitive\": false,\n                    \"type\": \"list\",\n                    \"value\": [\n                        \"some-env-bbl6-credhub\"\n                    ]\n                },\n                \"director_address\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"https://35.185.125.178:25555\"\n                },\n                \"external_ip\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"35.185.60.196\"\n                },\n                \"internal_tag_name\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"some-env-bbl6-internal\"\n                },\n                \"jumpbox_tag_name\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"some-env-bbl6-jumpbox\"\n                },\n                \"jumpbox_url\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"35.185.60.196:22\"\n                },\n                \"network_name\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"some-env-bbl6-network\"\n                },\n                \"router_backend_service\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"some-env-bbl6-router-lb\"\n                },\n                \"router_lb_ip\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"35.201.97.214\"\n                },\n                \"ssh_proxy_lb_ip\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"104.196.181.208\"\n                },\n                \"ssh_proxy_target_pool\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"some-env-bbl6-cf-ssh-proxy\"\n                },\n                \"subnetwork_name\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"some-env-bbl6-subnet\"\n                },\n                \"tcp_router_lb_ip\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"35.185.98.78\"\n                },\n                \"tcp_router_target_pool\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"some-env-bbl6-cf-tcp-router\"\n                },\n                \"ws_lb_ip\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"104.196.197.242\"\n                },\n                \"ws_target_pool\": {\n                    \"sensitive\": false,\n                    \"type\": \"string\",\n                    \"value\": \"some-env-bbl6-cf-ws\"\n                }\n            },\n            \"resources\": {},\n            \"depends_on\": []\n        }\n    ]\n}\n",
	"lb": {
		"type": "cf",
		"cert": "-----BEGIN CERTIFICATE-----\nsome-lb-cert\n-----END CERTIFICATE-----\n",
		"key": "-----BEGIN RSA PRIVATE KEY-----\nsome-lb-key\n-----END RSA PRIVATE KEY-----\n"
	},
	"latestTFOutput": "latest terraform error"
}
`
