package main_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("bbl", func() {
	Describe("unsupported-provision-aws-for-concourse", func() {
		Context("when the cloudformation stack does not exist", func() {
			It("creates and applies a cloudformation template", func() {
				tempDir, err := ioutil.TempDir("", "")

				state := storage.State{}

				buf, err := json.Marshal(state)
				Expect(err).NotTo(HaveOccurred())

				ioutil.WriteFile(filepath.Join(tempDir, "state.json"), buf, os.ModePerm)

				responses := []string{}

				var (
					callCount   int
					keyPairName string
				)
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.Method).To(Equal("POST"))

					body, err := ioutil.ReadAll(r.Body)
					Expect(err).NotTo(HaveOccurred())

					responses = append(responses, string(body))
					if strings.Contains(string(body), "Action=DescribeStack") {
						callCount++
						if callCount > 1 {
							w.Write([]byte(describeStacksResponse))
						}
						return
					}

					if strings.Contains(string(body), "Action=DescribeKeyPairs") {
						fmt.Fprint(w, describeKeyPairsResponse)
						return
					}

					if strings.Contains(string(body), "Action=CreateKeyPair") {
						values, err := url.ParseQuery(string(body))
						if err != nil {
							panic(err)
						}

						keyPairName = values.Get("KeyName")

						fmt.Fprintf(w, createKeyPairResponse, keyPairName)
						return
					}
				}))

				args := []string{
					fmt.Sprintf("--endpoint-override=%s", server.URL),
					"--aws-access-key-id", "some-access-key",
					"--aws-secret-access-key", "some-access-secret",
					"--aws-region", "some-region",
					"--state-dir", tempDir,
					"unsupported-provision-aws-for-concourse",
				}
				session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))

				Expect(responses).To(ContainElement(ContainSubstring("Action=CreateStack")))
				Expect(responses).To(ContainElement(ContainSubstring("StackName=concourse")))
				Expect(keyPairName).To(MatchRegexp(`keypair-\w{8}-\w{4}-\w{4}-\w{4}-\w{12}`))
			})
		})

		Context("when the cloudformation stack already exists", func() {
			It("updates the stack with the cloudformation template", func() {
				tempDir, err := ioutil.TempDir("", "")

				state := storage.State{
					KeyPair: &storage.KeyPair{
						Name: "some-keypair-name",
						PrivateKey: `-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----`,
					},
				}

				buf, err := json.Marshal(state)
				Expect(err).NotTo(HaveOccurred())

				ioutil.WriteFile(filepath.Join(tempDir, "state.json"), buf, os.ModePerm)

				responses := []string{}

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.Method).To(Equal("POST"))

					body, err := ioutil.ReadAll(r.Body)
					Expect(err).NotTo(HaveOccurred())
					responses = append(responses, string(body))
					if strings.Contains(string(body), "Action=DescribeStack") {
						w.Write([]byte(describeStacksResponse))
						return
					}

					if strings.Contains(string(body), "Action=DescribeKeyPairs") {
						fmt.Fprint(w, describeKeyPairsResponse)
						return
					}
				}))

				args := []string{
					fmt.Sprintf("--endpoint-override=%s", server.URL),
					"--aws-access-key-id", "some-access-key",
					"--aws-secret-access-key", "some-access-secret",
					"--aws-region", "some-region",
					"--state-dir", tempDir,
					"unsupported-provision-aws-for-concourse",
				}
				session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)

				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))

				Expect(responses).To(ContainElement(ContainSubstring("Action=UpdateStack")))
				Expect(responses).To(ContainElement(ContainSubstring("StackName=concourse")))
			})
		})
	})
})

var (
	describeStacksResponse = strings.TrimSpace(`
<DescribeStacksResponse xmlns="http://cloudformation.amazonaws.com/doc/2010-05-15/">
  <DescribeStacksResult>
    <Stacks>
      <member>
        <StackName>concourse</StackName>
        <StackId>arn:aws:cloudformation:us-east-1:123456789:stack/MyStack/aaf549a0-a413-11df-adb3-5081b3858e83</StackId>
        <CreationTime>2010-07-27T22:28:28Z</CreationTime>
        <StackStatus>CREATE_COMPLETE</StackStatus>
        <DisableRollback>false</DisableRollback>
        <Outputs>
          <member>
            <OutputKey>StartPage</OutputKey>
            <OutputValue>http://my-load-balancer.amazonaws.com:80/index.html</OutputValue>
          </member>
        </Outputs>
      </member>
    </Stacks>
  </DescribeStacksResult>
  <ResponseMetadata>
    <RequestId>b9b4b068-3a41-11e5-94eb-example</RequestId>
  </ResponseMetadata>
</DescribeStacksResponse>
`)

	describeKeyPairsResponse = strings.TrimSpace(`
<DescribeKeyPairsResponse xmlns="http://ec2.amazonaws.com/doc/2015-10-01/">
    <requestId>59dbff89-35bd-4eac-99ed-be587EXAMPLE</requestId> 
    <keySet>
      <item>
         <keyName>some-key-pair</keyName>
         <keyFingerprint>8a:85:cd:57:a4:ae:90:1e:6e:0d:af:4e:0f:e4:b6:df</keyFingerprint>
      </item>
   </keySet>
</DescribeKeyPairsResponse>	
`)

	createKeyPairResponse = strings.TrimSpace(`
<CreateKeyPairResponse xmlns="http://ec2.amazonaws.com/doc/2015-10-01/">
  <keyName>%s</keyName>
  <keyFingerprint>
     8a:85:cd:57:a4:ae:90:1e:6e:0d:af:4e:0f:e4:b6:df
  </keyFingerprint>
  <keyMaterial>-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----</keyMaterial>
</CreateKeyPairResponse>
`)
)
