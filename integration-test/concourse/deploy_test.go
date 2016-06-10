package integration_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pivotal-cf-experimental/bosh-bootloader/integration-test"
	"github.com/pivotal-cf-experimental/bosh-bootloader/integration-test/actors"
	"github.com/pivotal-cf-experimental/bosh-test/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("bosh deployment tests", func() {
	var (
		bbl   actors.BBL
		aws   actors.AWS
		state integration.State
	)

	BeforeEach(func() {
		var err error
		stateDirectory, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		configuration, err := integration.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(stateDirectory, pathToBBL, configuration)
		aws = actors.NewAWS(configuration)
		state = integration.NewState(stateDirectory)
	})

	It("is able to deploy concourse", func() {
		bbl.Up()

		bbl.CreateLB("concourse", "fixtures/bbl.crt", "fixtures/bbl.key", "")

		buf, err := ioutil.ReadFile("fixtures/concourse.yml")
		Expect(err).NotTo(HaveOccurred())

		concourseManifest := string(buf)

		boshClient := bosh.NewClient(bosh.Config{
			URL:              bbl.DirectorAddress(),
			Username:         bbl.DirectorUsername(),
			Password:         bbl.DirectorPassword(),
			AllowInsecureSSL: true,
		})

		lbURL := fmt.Sprintf("http://%s", aws.LoadBalancers(state.StackName())["ConcourseLoadBalancerURL"])

		info, err := boshClient.Info()
		Expect(err).NotTo(HaveOccurred())

		err = downloadAndUploadRelease(boshClient, "https://bosh.io/d/github.com/concourse/concourse")
		Expect(err).NotTo(HaveOccurred())

		err = downloadAndUploadRelease(boshClient, "https://bosh.io/d/github.com/cloudfoundry-incubator/garden-linux-release")
		Expect(err).NotTo(HaveOccurred())

		err = downloadAndUploadStemcell(boshClient, "https://bosh.io/d/stemcells/bosh-aws-xen-hvm-ubuntu-trusty-go_agent")
		Expect(err).NotTo(HaveOccurred())

		stemcell, err := boshClient.Stemcell("bosh-aws-xen-hvm-ubuntu-trusty-go_agent")
		Expect(err).NotTo(HaveOccurred())

		concourseRelease, err := boshClient.Release("concourse")
		Expect(err).NotTo(HaveOccurred())

		gardenLinuxRelease, err := boshClient.Release("garden-linux")
		Expect(err).NotTo(HaveOccurred())

		concourseManifest = strings.Replace(concourseManifest, "REPLACE_ME_DIRECTOR_UUID", info.UUID, -1)
		concourseManifest = strings.Replace(concourseManifest, "REPLACE_ME_EXTERNAL_URL", lbURL, -1)
		concourseManifest = strings.Replace(concourseManifest, "REPLACE_ME_STEMCELL_VERSION", stemcell.Latest(), -1)
		concourseManifest = strings.Replace(concourseManifest, "REPLACE_ME_CONCOURSE_VERSION", concourseRelease.Latest(), -1)
		concourseManifest = strings.Replace(concourseManifest, "REPLACE_ME_GARDEN_LINUX_VERSION", gardenLinuxRelease.Latest(), -1)

		boshClient.Deploy([]byte(concourseManifest))

		Eventually(func() ([]bosh.VM, error) {
			return boshClient.DeploymentVMs("concourse")
		}, "10m", "10s").Should(ConsistOf([]bosh.VM{
			{JobName: "worker", Index: 0, State: "running"},
			{JobName: "db", Index: 0, State: "running"},
			{JobName: "web", Index: 0, State: "running"},
		}))

		resp, err := http.Get(lbURL)
		Expect(err).NotTo(HaveOccurred())

		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		body, err := ioutil.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())

		Expect(string(body)).To(ContainSubstring("Log In - Concourse"))

		err = boshClient.DeleteDeployment("concourse")
		Expect(err).NotTo(HaveOccurred())

		bbl.Destroy()
	})
})

func downloadAndUploadStemcell(boshClient bosh.Client, stemcell string) error {
	file, size, err := download(stemcell)
	if err != nil {
		return err
	}

	_, err = boshClient.UploadStemcell(bosh.NewSizeReader(file, size))
	return err
}

func downloadAndUploadRelease(boshClient bosh.Client, release string) error {
	file, size, err := download(release)
	if err != nil {
		return err
	}

	_, err = boshClient.UploadRelease(bosh.NewSizeReader(file, size))
	return err
}

func download(location string) (io.Reader, int64, error) {
	resp, err := http.Get(location)
	if err != nil {
		return nil, 0, err
	}

	return resp.Body, resp.ContentLength, nil
}
