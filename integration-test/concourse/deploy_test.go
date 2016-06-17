package integration_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pivotal-cf-experimental/bosh-bootloader/integration-test"
	"github.com/pivotal-cf-experimental/bosh-bootloader/integration-test/actors"
	"github.com/pivotal-cf-experimental/bosh-bootloader/testhelpers"
	"github.com/pivotal-cf-experimental/bosh-test/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	ConcourseExampleManifestURL = "https://raw.githubusercontent.com/concourse/concourse/develop/manifests/concourse.yml"
	ConcourseReleaseURL         = "https://bosh.io/d/github.com/concourse/concourse"
	GardenReleaseURL            = "https://bosh.io/d/github.com/cloudfoundry-incubator/garden-runc-release"
	GardenReleaseName           = "garden-runc"
	StemcellURL                 = "https://bosh.io/d/stemcells/bosh-aws-xen-hvm-ubuntu-trusty-go_agent"
	StemcellName                = "bosh-aws-xen-hvm-ubuntu-trusty-go_agent"
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

		certPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_CERT)
		Expect(err).NotTo(HaveOccurred())

		keyPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_KEY)
		Expect(err).NotTo(HaveOccurred())

		bbl.CreateLB("concourse", certPath, keyPath, "")

		boshClient := bosh.NewClient(bosh.Config{
			URL:              bbl.DirectorAddress(),
			Username:         bbl.DirectorUsername(),
			Password:         bbl.DirectorPassword(),
			AllowInsecureSSL: true,
		})

		err = downloadAndUploadRelease(boshClient, ConcourseReleaseURL)
		Expect(err).NotTo(HaveOccurred())

		err = downloadAndUploadRelease(boshClient, GardenReleaseURL)
		Expect(err).NotTo(HaveOccurred())

		err = downloadAndUploadStemcell(boshClient, StemcellURL)
		Expect(err).NotTo(HaveOccurred())

		concourseExampleManifest, err := downloadConcourseExampleManifest()
		Expect(err).NotTo(HaveOccurred())

		info, err := boshClient.Info()
		Expect(err).NotTo(HaveOccurred())

		lbURL := fmt.Sprintf("http://%s", aws.LoadBalancers(state.StackName())["ConcourseLoadBalancerURL"])

		stemcell, err := boshClient.Stemcell(StemcellName)
		Expect(err).NotTo(HaveOccurred())

		concourseRelease, err := boshClient.Release("concourse")
		Expect(err).NotTo(HaveOccurred())

		gardenRelease, err := boshClient.Release(GardenReleaseName)
		Expect(err).NotTo(HaveOccurred())

		concourseManifestInputs := concourseManifestInputs{
			boshDirectorUUID:        info.UUID,
			webExternalURL:          lbURL,
			stemcellVersion:         stemcell.Latest(),
			concourseReleaseVersion: concourseRelease.Latest(),
			gardenReleaseVersion:    gardenRelease.Latest(),
		}
		concourseManifest, err := populateManifest(concourseExampleManifest, concourseManifestInputs)
		Expect(err).NotTo(HaveOccurred())

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

func downloadConcourseExampleManifest() (string, error) {
	resp, _, err := download(ConcourseExampleManifestURL)
	if err != nil {
		return "", err
	}

	rawRespBody, err := ioutil.ReadAll(resp)
	if err != nil {
		return "", err
	}

	return string(rawRespBody), nil
}

func download(location string) (io.Reader, int64, error) {
	resp, err := http.Get(location)
	if err != nil {
		return nil, 0, err
	}

	return resp.Body, resp.ContentLength, nil
}
