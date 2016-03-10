package main_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"
	"github.com/pivotal-cf-experimental/bosh-bootloader/bbl/awsbackend"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
	"github.com/rosenhouse/awsfaker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

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

var _ = Describe("bbl", func() {
	var (
		fakeAWS       *awsbackend.Backend
		server        *httptest.Server
		tempDirectory string
	)

	BeforeEach(func() {
		fakeAWS = awsbackend.New()
		server = httptest.NewServer(awsfaker.New(fakeAWS))

		var err error
		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("unsupported-deploy-bosh-on-aws-for-concourse", func() {
		Context("when the cloudformation stack does not exist", func() {
			var stack awsbackend.Stack

			BeforeEach(func() {
				writeEmptyStateJson(tempDirectory)
			})

			It("creates a stack and a keypair", func() {
				deployBOSHOnAWSForConcourse(server.URL, tempDirectory)

				var ok bool
				stack, ok = fakeAWS.Stacks.Get("concourse")
				Expect(ok).To(BeTrue())

				Expect(stack.Name).To(Equal("concourse"))

				keyPairs := fakeAWS.KeyPairs.All()
				Expect(keyPairs).To(HaveLen(1))
				Expect(keyPairs[0].Name).To(MatchRegexp(`keypair-\w{8}-\w{4}-\w{4}-\w{4}-\w{12}`))
			})

			It("creates an IAM user", func() {
				deployBOSHOnAWSForConcourse(server.URL, tempDirectory)

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
				session := deployBOSHOnAWSForConcourse(server.URL, tempDirectory)

				stdout := session.Out.Contents()
				Expect(stdout).To(ContainSubstring("step: creating keypair"))
				Expect(stdout).To(ContainSubstring("step: generating cloudformation template"))
				Expect(stdout).To(ContainSubstring("step: creating cloudformation stack"))
				Expect(stdout).To(ContainSubstring("step: finished applying cloudformation template"))
				Expect(stdout).To(ContainSubstring("step: generating bosh-init manifest"))

				Expect(stdout).To(ContainSubstring("bosh-init manifest:"))
				Expect(stdout).To(ContainSubstring("name: bosh"))
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

				session := deployBOSHOnAWSForConcourse(server.URL, tempDirectory)

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
	})
})

func writeEmptyStateJson(tempDirectory string) {
	state := storage.State{}

	buf, err := json.Marshal(state)
	Expect(err).NotTo(HaveOccurred())

	ioutil.WriteFile(filepath.Join(tempDirectory, "state.json"), buf, os.ModePerm)
}

func deployBOSHOnAWSForConcourse(serverURL string, tempDirectory string) *gexec.Session {
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
	Eventually(session, 10*time.Second).Should(gexec.Exit(0))

	return session
}
