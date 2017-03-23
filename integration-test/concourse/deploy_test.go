package integration_test

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	integration "github.com/cloudfoundry/bosh-bootloader/integration-test"
	"github.com/cloudfoundry/bosh-bootloader/integration-test/actors"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf-experimental/bosh-test/bosh"
)

const (
	ConcourseExampleManifestURL = "https://raw.githubusercontent.com/concourse/concourse/master/docs/setting-up/installing.any"
	ConcourseReleaseURL         = "https://bosh.io/d/github.com/concourse/concourse"
	GardenReleaseURL            = "https://bosh.io/d/github.com/cloudfoundry/garden-runc-release"
	GardenReleaseName           = "garden-runc"
	AWSStemcellURL              = "https://bosh.io/d/stemcells/bosh-aws-xen-hvm-ubuntu-trusty-go_agent"
	AWSStemcellName             = "bosh-aws-xen-hvm-ubuntu-trusty-go_agent"
	GCPStemcellURL              = "https://bosh.io/d/stemcells/bosh-google-kvm-ubuntu-trusty-go_agent"
	GCPStemcellName             = "bosh-google-kvm-ubuntu-trusty-go_agent"
)

var _ = Describe("concourse deployment test", func() {
	var deployConcourseTest = func(bbl actors.BBL, stemcellURL, stemcellName, lbURL string, tlsMode bool, tlsBindPort int) {
		boshClient := bosh.NewClient(bosh.Config{
			URL:              bbl.DirectorAddress(),
			Username:         bbl.DirectorUsername(),
			Password:         bbl.DirectorPassword(),
			AllowInsecureSSL: true,
		})

		err := downloadAndUploadRelease(boshClient, ConcourseReleaseURL)
		Expect(err).NotTo(HaveOccurred())

		err = downloadAndUploadRelease(boshClient, GardenReleaseURL)
		Expect(err).NotTo(HaveOccurred())

		err = downloadAndUploadStemcell(boshClient, stemcellURL)
		Expect(err).NotTo(HaveOccurred())

		concourseExampleManifest, err := downloadConcourseExampleManifest()
		Expect(err).NotTo(HaveOccurred())

		info, err := boshClient.Info()
		Expect(err).NotTo(HaveOccurred())

		stemcell, err := boshClient.StemcellByName(stemcellName)
		Expect(err).NotTo(HaveOccurred())

		concourseRelease, err := boshClient.Release("concourse")
		Expect(err).NotTo(HaveOccurred())

		gardenRelease, err := boshClient.Release(GardenReleaseName)
		Expect(err).NotTo(HaveOccurred())

		stemcellLatest, err := stemcell.Latest()
		Expect(err).NotTo(HaveOccurred())

		concourseManifestInputs := concourseManifestInputs{
			boshDirectorUUID:        info.UUID,
			webExternalURL:          lbURL,
			tlsMode:                 tlsMode,
			tlsBindPort:             tlsBindPort,
			stemcellVersion:         stemcellLatest,
			concourseReleaseVersion: concourseRelease.Latest(),
			gardenReleaseVersion:    gardenRelease.Latest(),
		}
		concourseManifest, err := populateManifest(concourseExampleManifest, concourseManifestInputs)
		Expect(err).NotTo(HaveOccurred())

		_, err = boshClient.Deploy([]byte(concourseManifest))
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() ([]bosh.VM, error) {
			vms, err := boshClient.DeploymentVMs("concourse")
			if err != nil {
				return []bosh.VM{}, err
			}

			vmsNoID := []bosh.VM{}
			for _, vm := range vms {
				vm.ID = ""
				vmsNoID = append(vmsNoID, vm)
			}
			return vmsNoID, nil
		}, "1m", "10s").Should(ConsistOf([]bosh.VM{
			{JobName: "worker", Index: 0, State: "running"},
			{JobName: "db", Index: 0, State: "running"},
			{JobName: "web", Index: 0, State: "running"},
		}))

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}

		resp, err := client.Get(lbURL)
		Expect(err).NotTo(HaveOccurred())

		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		body, err := ioutil.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())

		Expect(string(body)).To(ContainSubstring("<title>Concourse</title>"))

		err = boshClient.DeleteDeployment("concourse")
		Expect(err).NotTo(HaveOccurred())

		bbl.Destroy()
	}

	Describe("aws", func() {
		var (
			bbl   actors.BBL
			aws   actors.AWS
			state integration.State
			lbURL string
		)

		BeforeEach(func() {
			var err error
			configuration, err := integration.LoadAWSConfig()
			Expect(err).NotTo(HaveOccurred())

			bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "concourse-env")
			aws = actors.NewAWS(configuration)
			state = integration.NewState(configuration.StateFileDir)

			bbl.Up(actors.AWSIAAS, []string{"--name", bbl.PredefinedEnvID()})

			certPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_CERT)
			Expect(err).NotTo(HaveOccurred())

			keyPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_KEY)
			Expect(err).NotTo(HaveOccurred())

			bbl.CreateLB("concourse", certPath, keyPath, "")

			lbURL = fmt.Sprintf("http://%s", aws.LoadBalancers(state.StackName())["ConcourseLoadBalancerURL"])
		})

		It("is able to deploy concourse", func() {
			deployConcourseTest(bbl, AWSStemcellURL, AWSStemcellName, lbURL, false, 0)
		})
	})

	Describe("gcp", func() {
		var (
			bbl   actors.BBL
			gcp   actors.GCP
			state integration.State
			lbURL string
		)

		BeforeEach(func() {
			var err error
			configuration, err := integration.LoadGCPConfig()
			Expect(err).NotTo(HaveOccurred())

			bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "concourse-env")
			gcp = actors.NewGCP(configuration)
			state = integration.NewState(configuration.StateFileDir)

			bbl.Up(actors.GCPIAAS, []string{"--name", bbl.PredefinedEnvID()})

			bbl.CreateGCPLB("concourse")

			envID := bbl.EnvID()
			address, err := gcp.GetAddress(envID + "-concourse")
			Expect(err).NotTo(HaveOccurred())

			lbURL = fmt.Sprintf("https://%s", address.Address)
		})

		It("is able to deploy concourse", func() {
			deployConcourseTest(bbl, GCPStemcellURL, GCPStemcellName, lbURL, true, 443)
		})
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

	var lines []string
	scanner := bufio.NewScanner(resp)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	startIndexOfYamlCode := -1
	endIndexOfYamlCode := -1

	for index, line := range lines {
		startMatched, startErr := regexp.MatchString(`^\s*\\codeblock{yaml}{$`, line)
		endMatched, endErr := regexp.MatchString(`^\s*}$`, line)
		if endErr != nil {
			return "", endErr
		}

		if startErr != nil {
			panic(startErr)
		}

		if startMatched && startIndexOfYamlCode < 0 {
			startIndexOfYamlCode = index + 1
		}
		if endMatched && endIndexOfYamlCode < 0 && startIndexOfYamlCode > 0 {
			endIndexOfYamlCode = index
		}
	}

	yamlDocument := lines[startIndexOfYamlCode:endIndexOfYamlCode]

	re := regexp.MustCompile(`^(\s*)---`)
	results := re.FindAllStringSubmatch(yamlDocument[0], -1)
	indentation := results[0][1]
	for index, line := range yamlDocument {
		indentationRegexp := regexp.MustCompile(fmt.Sprintf(`^%s`, indentation))
		escapesRegexp := regexp.MustCompile(`\\([{}])`)
		tlsRegexp := regexp.MustCompile("^.*(tls_key|tls_cert).*$")

		line = indentationRegexp.ReplaceAllString(line, "")
		line = escapesRegexp.ReplaceAllString(line, "$1")
		line = tlsRegexp.ReplaceAllString(line, "")

		yamlDocument[index] = line

	}

	yamlString := strings.Join(yamlDocument, "\n")
	return yamlString, nil
}

func download(location string) (io.Reader, int64, error) {
	resp, err := http.Get(location)
	if err != nil {
		return nil, 0, err
	}

	return resp.Body, resp.ContentLength, nil
}
