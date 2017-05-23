package integration_test

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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
)

var _ = Describe("concourse deployment test", func() {
	var (
		bbl           actors.BBL
		state         integration.State
		lbURL         string
		configuration integration.Config
	)

	BeforeEach(func() {
		var err error
		configuration, err = integration.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "concourse-env")
		state = integration.NewState(configuration.StateFileDir)

		fmt.Printf("using state-dir: %s\n", configuration.StateFileDir)
		bbl.Up(actors.GetIAAS(configuration), []string{"--name", bbl.PredefinedEnvID()})

		certPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_CERT)
		Expect(err).NotTo(HaveOccurred())

		keyPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_KEY)
		Expect(err).NotTo(HaveOccurred())

		bbl.CreateLB("concourse", certPath, keyPath, "")

		lbURL, err = actors.LBURL(configuration, bbl, state)
		Expect(err).NotTo(HaveOccurred())
	})

	It("is able to deploy concourse", func() {
		boshClient := bosh.NewClient(bosh.Config{
			URL:              bbl.DirectorAddress(),
			Username:         bbl.DirectorUsername(),
			Password:         bbl.DirectorPassword(),
			AllowInsecureSSL: true,
		})

		err := uploadRelease(boshClient, configuration.ConcourseReleasePath)
		Expect(err).NotTo(HaveOccurred())

		err = uploadRelease(boshClient, configuration.GardenReleasePath)
		Expect(err).NotTo(HaveOccurred())

		err = uploadStemcell(boshClient, configuration.StemcellPath)
		Expect(err).NotTo(HaveOccurred())

		concourseExampleManifest, err := downloadConcourseExampleManifest()
		Expect(err).NotTo(HaveOccurred())

		stemcell, err := boshClient.StemcellByName(configuration.StemcellName)
		Expect(err).NotTo(HaveOccurred())

		concourseRelease, err := boshClient.Release("concourse")
		Expect(err).NotTo(HaveOccurred())

		gardenRelease, err := boshClient.Release("garden-runc")
		Expect(err).NotTo(HaveOccurred())

		stemcellLatest, err := stemcell.Latest()
		Expect(err).NotTo(HaveOccurred())

		tlsMode := false
		tlsBindPort := 0
		switch actors.GetIAAS(configuration) {
		case actors.GCPIAAS:
			tlsMode = true
			tlsBindPort = 443
		}

		concourseManifestInputs := concourseManifestInputs{
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
				vm.IPs = nil
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
	})
})

func uploadStemcell(boshClient bosh.Client, stemcellPath string) error {
	stemcell, err := openFile(stemcellPath)
	if err != nil {
		return err
	}

	_, err = boshClient.UploadStemcell(stemcell)
	return err
}

func uploadRelease(boshClient bosh.Client, releasePath string) error {
	release, err := openFile(releasePath)
	if err != nil {
		return err
	}

	_, err = boshClient.UploadRelease(release)
	return err
}

func openFile(filePath string) (bosh.SizeReader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	return bosh.NewSizeReader(file, stat.Size()), nil
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
