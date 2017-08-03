package acceptance_test

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"

	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf-experimental/bosh-test/bosh"
)

var _ = Describe("concourse deployment test", func() {
	var (
		bbl           actors.BBL
		state         acceptance.State
		lbURL         string
		configuration acceptance.Config
	)

	BeforeEach(func() {
		var err error
		configuration, err = acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "concourse-env")
		state = acceptance.NewState(configuration.StateFileDir)

		bbl.Up(configuration.IAAS, []string{"--name", bbl.PredefinedEnvID()})

		certPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_CERT)
		Expect(err).NotTo(HaveOccurred())

		keyPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_KEY)
		Expect(err).NotTo(HaveOccurred())

		bbl.CreateLB("concourse", certPath, keyPath, "")

		lbURL, err = actors.LBURL(configuration, bbl, state)
		Expect(err).NotTo(HaveOccurred())
	})

	It("is able to deploy concourse and teardown infrastructure", func() {
		boshClient := bosh.NewClient(bosh.Config{
			URL:              bbl.DirectorAddress(),
			Username:         bbl.DirectorUsername(),
			Password:         bbl.DirectorPassword(),
			AllowInsecureSSL: true,
		})

		By("uploading releases and stemcells", func() {
			err := uploadRelease(boshClient, configuration.ConcourseReleasePath)
			Expect(err).NotTo(HaveOccurred())

			err = uploadRelease(boshClient, configuration.GardenReleasePath)
			Expect(err).NotTo(HaveOccurred())

			err = uploadStemcell(boshClient, configuration.StemcellPath)
			Expect(err).NotTo(HaveOccurred())
		})

		By("running bosh deploy and checking all the vms are running", func() {
			os.Setenv("BOSH_CLIENT", bbl.DirectorUsername())
			os.Setenv("BOSH_CLIENT_SECRET", bbl.DirectorPassword())
			os.Setenv("BOSH_ENVIRONMENT", bbl.DirectorAddress())
			os.Setenv("BOSH_CA_CERT", bbl.DirectorCACert())
			args := []string{
				"-d", "concourse",
				"deploy",
				fmt.Sprintf("%s/concourse-deployment.yml", configuration.ConcourseDeploymentPath),
				"-o", fmt.Sprintf("%s/operations/%s.yml", configuration.ConcourseDeploymentPath, actors.IAASString(configuration)),
				"--vars-store", "concourse-vars.yml",
				"-v", fmt.Sprintf("domain=%s", lbURL),
				"-n",
			}
			cmd := exec.Command("bosh", args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
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
		})

		By("testing the deployment", func() {
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
		})

		By("deleting the deployment", func() {
			err := boshClient.DeleteDeployment("concourse")
			Expect(err).NotTo(HaveOccurred())
		})

		By("deleting load balancers", func() {
			bbl.DeleteLBs()
		})

		By("tearing down the infrastructure", func() {
			bbl.Destroy()
		})
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
