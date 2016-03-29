package main_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"
	"github.com/pivotal-cf-experimental/bosh-bootloader/bbl/awsbackend"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
	"github.com/rosenhouse/awsfaker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

const expectedCloudConfig = `
azs:
- cloud_properties:
    availability_zone: us-east-1a
  name: z1
- cloud_properties:
    availability_zone: us-east-1b
  name: z2
- cloud_properties:
    availability_zone: us-east-1c
  name: z3
- cloud_properties:
    availability_zone: us-east-1e
  name: z4
compilation:
  az: z1
  network: concourse
  reuse_compilation_vms: true
  vm_type: default
  workers: 3
disk_types:
- cloud_properties:
    type: gp2
  disk_size: 1024
  name: default
networks:
- name: concourse
  subnets:
  - az: z1
    cloud_properties:
      security_groups:
      - some-security-group-1
      subnet: some-subnet-1
    gateway: 10.0.16.1
    range: 10.0.16.0/20
    reserved:
    - 10.0.16.2-10.0.16.3
    - 10.0.31.255
    static: []
  - az: z2
    cloud_properties:
      security_groups:
      - some-security-group-2
      subnet: some-subnet-2
    gateway: 10.0.32.1
    range: 10.0.32.0/20
    reserved:
    - 10.0.32.2-10.0.32.3
    - 10.0.47.255
    static: []
  - az: z3
    cloud_properties:
      security_groups:
      - some-security-group-3
      subnet: some-subnet-3
    gateway: 10.0.48.1
    range: 10.0.48.0/20
    reserved:
    - 10.0.48.2-10.0.48.3
    - 10.0.63.255
    static: []
  type: manual
vm_types:
- cloud_properties:
    ephemeral_disk:
      size: 1024
      type: gp2
    instance_type: m3.medium
  name: m3.medium
- cloud_properties:
    ephemeral_disk:
      size: 1024
      type: gp2
    instance_type: m3.large
  name: m3.large
- cloud_properties:
    ephemeral_disk:
      size: 1024
      type: gp2
    instance_type: c3.large
  name: c3.large
- cloud_properties:
    ephemeral_disk:
      size: 1024
      type: gp2
    instance_type: c3.xlarge
  name: c3.xlarge
- cloud_properties:
    ephemeral_disk:
      size: 1024
      type: gp2
    instance_type: c3.2xlarge
  name: c3.2xlarge
- cloud_properties:
    ephemeral_disk:
      size: 1024
      type: gp2
    instance_type: c4.large
  name: c4.large
- cloud_properties:
    ephemeral_disk:
      size: 1024
      type: gp2
    instance_type: r3.xlarge
  name: r3.xlarge
- cloud_properties:
    ephemeral_disk:
      size: 1024
      type: gp2
    instance_type: t2.micro
  name: t2.micro`

const privateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAt5oGrrqGwYvxJT3L37olM4X67ZNnWt7IXNTc0c61wzlyPkvU
ReUoVDtxkuD6iNaU1AiVXxZ5xwqCdbxk+pH2y0bini7W50TEoVxNllJwKDU32c2L
UyKLfyPVijafae90Mtuilo8Pyyl3xqs2JKs07IjA3rIwLzom1SEu7LuO3eeuMeyw
T4cy3J3zRRYP2eEZ8IZ4WkMv0Pgkn7t696dIcV+U89xyze/WW0y8QOeTFMkDIcpg
lFfrvSmxN4kV/+LJaJnQqfk8rTnySYgT6Yeod9mjdNx4LseYL2HMLSm4UO9YF21D
cKQH324zlsB71kDn6b/riLgY09vBZhDj/E0uHwIDAQABAoIBACP7f8vGqppL/tq5
nbcfGCNc4qyk8uCQQNxQq2ZDCMRWAdnLqrJ4EstPSxbqGK+wvkI/3GZiVUN4/9Br
N68T5DY6kjdGHr/8bjzhhiMrzOdUZrm82s1UO9qS/0qzIdL1JuTAvsCbERFT8zFw
ZJATLbAdrQ74BRF8aBflBPlIWNuMMx/nFV+GkOgRq1xvVdPYqtimT3cs/4Akuf9o
LZZQZp4eSEJJp+JVGQpmOMak9dbpjyU8znWf69qrN6E7kfPfXl1csX2N1eV0nJq0
4uuyUUsG04zIE2JWu8MW0pLDLDD8Nw56BZ6Zo7g1R0KYyXguSi079sEBRHS5fiVx
HAP8DYECgYEA591z08bTt9Lm+SulXEadWwMwLlVnAXCGkKpTHnTZg2Q64awSi9Xq
i7UhsR6DRhgrxgf07dAm0mgLWHmN834JP0jMmy/Pm/+1ck3edq6SMadQMrTdgMJD
Z2cQW4W86MQ7Z3L+nxIYVDypKYQk7CxmVCRvHRzCqPcyJShJfaHaPHECgYEAyrZ9
swZFSm6tGi/ehMrdFbjvHxjIRque5juuvVQLdYOtitkXRdfKtJ1puXpLfw6k7lsM
8Y/YGGdk8CH5KGsFlVncYTOyxGi+a21m2ePfmXf1f1j/XKCx1ObhoZL6+6KKKawk
5MaF6kp+QNjOL5MOl14v9aCoO652XnmWlBgdm48CgYBTxki2SM1wSoxXlPR/PahX
HPTImOTJuV11YYT8qR16ArnfletxiM3gwoY016B4r/0I5RES57VPKnaG9gxa4Lv4
mJYMsB6j76UgcpAhc3uw4xHv8Ddj8UynTK61UsHpnBUWkI787G3L6cr5DBzHFFe4
qR1YeG7A2+fLUx4SfWs7kQKBgHOPv278pym8mIAyQ+duAsVsbR1MMnhfRDG6Wm5i
aDnw/FEIW4UcdNmsV2Y+eqWPQqUDUQiw2R9oahmfNHw/Lqqq1MCxCTuA/vUdJCIZ
DxJdWZ3krYcvsNFPYdeLg/tJ+PuywEGPjy42k20Ca+ChNBNExZCAqweC+MX5CMea
S96vAoGBAKBP0opR+RiJ9cW7Aol8KaGZdk8tSehudgTchkqXyqfTOqnkLWCprQuN
O9wJ7sJLZLyHhV+ENrBZFashTJetQAPVT3ziwvasJq566g1y+Db3/8HAzOZd9toT
ohmMhda49PmtPpDlTAMihjbjvLAM7IU/S7+FVIINjTBV+YVnjS2y
-----END RSA PRIVATE KEY-----`

type fakeBOSHDirector struct {
	CloudConfig []byte
}

func (b *fakeBOSHDirector) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	buf, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}
	b.CloudConfig = buf
	responseWriter.WriteHeader(http.StatusCreated)
}

var _ = Describe("bbl", func() {
	var (
		fakeAWS        *awsbackend.Backend
		fakeAWSServer  *httptest.Server
		fakeBOSHServer *httptest.Server
		tempDirectory  string
		fakeBOSH       *fakeBOSHDirector
	)

	BeforeEach(func() {
		fakeBOSH = &fakeBOSHDirector{}
		fakeBOSHServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			fakeBOSH.ServeHTTP(responseWriter, request)
		}))

		fakeAWS = awsbackend.New(fakeBOSHServer.URL)
		fakeAWSServer = httptest.NewServer(awsfaker.New(fakeAWS))

		var err error
		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("unsupported-deploy-bosh-on-aws-for-concourse", func() {
		Context("when the cloudformation stack does not exist", func() {
			var stack awsbackend.Stack

			It("creates a stack and a keypair", func() {
				deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)

				var ok bool
				stack, ok = fakeAWS.Stacks.Get("concourse")
				Expect(ok).To(BeTrue())

				Expect(stack.Name).To(Equal("concourse"))

				keyPairs := fakeAWS.KeyPairs.All()
				Expect(keyPairs).To(HaveLen(1))
				Expect(keyPairs[0].Name).To(MatchRegexp(`keypair-\w{8}-\w{4}-\w{4}-\w{4}-\w{12}`))
			})

			It("creates an IAM user", func() {
				deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)

				var ok bool
				stack, ok = fakeAWS.Stacks.Get("concourse")
				Expect(ok).To(BeTrue())

				var template struct {
					Resources struct {
						BOSHUser struct {
							Properties templates.IAMUser
							Type       string
						}
					}
				}

				err := json.Unmarshal([]byte(stack.Template), &template)
				Expect(err).NotTo(HaveOccurred())

				Expect(template.Resources.BOSHUser.Properties.Policies).To(HaveLen(1))
			})

			It("logs the steps and bosh-init manifest", func() {
				session := deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)

				stdout := session.Out.Contents()
				Expect(stdout).To(ContainSubstring("step: creating keypair"))
				Expect(stdout).To(ContainSubstring("step: generating cloudformation template"))
				Expect(stdout).To(ContainSubstring("step: creating cloudformation stack"))
				Expect(stdout).To(ContainSubstring("step: finished applying cloudformation template"))
				Expect(stdout).To(ContainSubstring("step: generating bosh-init manifest"))
				Expect(stdout).To(ContainSubstring("step: deploying bosh director"))
				Expect(stdout).To(ContainSubstring("Director Address:  127.0.0.1"))
			})

			It("prints out randomized bosh director credentials", func() {
				session := deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)

				stdout := session.Out.Contents()
				Expect(stdout).To(MatchRegexp(`Director Username: user-\w{7}`))
				Expect(stdout).To(MatchRegexp(`Director Password: p-\w{15}`))
			})

			It("invokes bosh-init", func() {
				session := deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)
				Expect(session.Out.Contents()).To(ContainSubstring("bosh-init was called with [bosh-init deploy bosh.yml]"))
				Expect(session.Out.Contents()).To(ContainSubstring("bosh-state.json: {}"))
			})

			It("can invoke bosh-init idempotently", func() {
				session := deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)
				Expect(session.Out.Contents()).To(ContainSubstring("bosh-init was called with [bosh-init deploy bosh.yml]"))
				Expect(session.Out.Contents()).To(ContainSubstring("bosh-state.json: {}"))

				session = deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)
				Expect(session.Out.Contents()).To(ContainSubstring("bosh-init was called with [bosh-init deploy bosh.yml]"))
				Expect(session.Out.Contents()).To(ContainSubstring(`bosh-state.json: {"key":"value","md5checksum":`))
				Expect(session.Out.Contents()).To(ContainSubstring("No new changes, skipping deployment..."))
			})

			It("fast fails if the bosh state exists", func() {
				writeStateJson(storage.State{BOSH: &storage.BOSH{}}, tempDirectory)
				session := deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 1)
				Expect(session.Err.Contents()).To(ContainSubstring("Found BOSH data in state directory"))
			})
		})

		Context("when the keypair and cloudformation stack already exist", func() {
			BeforeEach(func() {
				fakeAWS.Stacks.Set(awsbackend.Stack{
					Name: "concourse",
				})
				fakeAWS.KeyPairs.Set(awsbackend.KeyPair{
					Name: "some-keypair-name",
				})
			})

			It("updates the stack with the cloudformation template", func() {
				state := storage.State{
					KeyPair: &storage.KeyPair{
						Name:       "some-keypair-name",
						PrivateKey: privateKey,
					},
				}

				buf, err := json.Marshal(state)
				Expect(err).NotTo(HaveOccurred())

				ioutil.WriteFile(filepath.Join(tempDirectory, "state.json"), buf, os.ModePerm)

				session := deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)

				stack, ok := fakeAWS.Stacks.Get("concourse")
				Expect(ok).To(BeTrue())
				Expect(stack).To(Equal(awsbackend.Stack{
					Name:       "concourse",
					WasUpdated: true,
				}))

				stdout := session.Out.Contents()
				Expect(stdout).To(ContainSubstring("step: using existing keypair"))
				Expect(stdout).To(ContainSubstring("step: generating cloudformation template"))
				Expect(stdout).To(ContainSubstring("step: updating cloudformation stack"))
				Expect(stdout).To(ContainSubstring("step: finished applying cloudformation template"))
			})
		})

		Context("cloud config", func() {
			It("applies the cloud config", func() {
				session := deployBOSHOnAWSForConcourse(fakeAWSServer.URL, tempDirectory, 0)
				stdout := session.Out.Contents()
				Expect(stdout).To(ContainSubstring("step: generating cloud config"))
				Expect(stdout).To(ContainSubstring("step: applying cloud config"))
				Expect(fakeBOSH.CloudConfig).To(MatchYAML(expectedCloudConfig))
			})
		})
	})
})

func writeStateJson(state storage.State, tempDirectory string) {
	buf, err := json.Marshal(state)
	Expect(err).NotTo(HaveOccurred())

	ioutil.WriteFile(filepath.Join(tempDirectory, "state.json"), buf, os.ModePerm)
}

func deployBOSHOnAWSForConcourse(serverURL string, tempDirectory string, exitCode int) *gexec.Session {
	args := []string{
		fmt.Sprintf("--endpoint-override=%s", serverURL),
		"--aws-access-key-id", "some-access-key",
		"--aws-secret-access-key", "some-access-secret",
		"--aws-region", "some-region",
		"--state-dir", tempDirectory,
		"unsupported-deploy-bosh-on-aws-for-concourse",
	}

	session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, 10*time.Second).Should(gexec.Exit(exitCode))

	return session
}
