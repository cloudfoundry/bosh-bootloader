//go:generate packr2

package bosh_test

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/fileio"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/gobuffalo/packr/v2"
	"github.com/spf13/afero"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Executor", func() {
	var (
		fs                    *afero.Afero
		cli                   *fakes.BOSHCLI
		stateDir              string
		deploymentDir         string
		varsDir               string
		relativeDeploymentDir string
		relativeVarsDir       string
		relativeStateDir      string
		dirInput              bosh.DirInput

		executor bosh.Executor
	)

	BeforeEach(func() {
		fs = &afero.Afero{Fs: afero.NewMemMapFs()}
		cli = &fakes.BOSHCLI{}
		cli.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
			stdout.Write([]byte("some-manifest"))
			return nil
		}
		cli.GetBOSHPathCall.Returns.Path = "bosh-path"

		var err error
		stateDir, err = fs.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		deploymentDir = filepath.Join(stateDir, "deployment")
		err = fs.Mkdir(deploymentDir, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		varsDir = filepath.Join(stateDir, "vars")
		err = fs.Mkdir(varsDir, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		relativeDeploymentDir = "${BBL_STATE_DIR}/deployment"
		relativeVarsDir = "${BBL_STATE_DIR}/vars"
		relativeStateDir = "${BBL_STATE_DIR}"

		dirInput = bosh.DirInput{
			VarsDir:  varsDir,
			StateDir: stateDir,
		}

		box := packr.New("some-name", "./deployments")
		box.AddString("jumpbox-deployment/no-external-ip.yml", "no-ip")
		box.AddString("jumpbox-deployment/aws/cpi.yml", "vsphere-cpi")
		box.AddString("jumpbox-deployment/azure/cpi.yml", "azure-cpi")
		box.AddString("bosh-deployment/vsphere/cpi.yml", "vsphere-cpi")
		box.AddString("bosh-deployment/LICENSE", "my-license")

		executor = bosh.Executor{CLI: cli, FS: fs, Box: box}
	})

	Describe("PlanJumpbox", func() {
		It("writes bosh-deployment assets to the deployment dir", func() {
			err := executor.PlanJumpbox(dirInput, deploymentDir, "aws")
			Expect(err).NotTo(HaveOccurred())

			By("writing bosh-deployment assets to the deployment dir", func() {
				simplePath := filepath.Join(deploymentDir, "no-external-ip.yml")

				contents, err := fs.ReadFile(simplePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal("no-ip"))

				nestedPath := filepath.Join(deploymentDir, "aws", "cpi.yml")

				contents, err = fs.ReadFile(nestedPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal("vsphere-cpi"))
			})

			By("writing create-env and delete-env scripts", func() {
				expectedArgs := []string{
					fmt.Sprintf("%s/jumpbox.yml", relativeDeploymentDir),
					"--state", fmt.Sprintf("%s/jumpbox-state.json", relativeVarsDir),
					"--vars-store", fmt.Sprintf("%s/jumpbox-vars-store.yml", relativeVarsDir),
					"--vars-file", fmt.Sprintf("%s/jumpbox-vars-file.yml", relativeVarsDir),
					"-o", fmt.Sprintf("%s/aws/cpi.yml", relativeDeploymentDir),
					"-v", `access_key_id="${BBL_AWS_ACCESS_KEY_ID}"`,
					"-v", `secret_access_key="${BBL_AWS_SECRET_ACCESS_KEY}"`,
				}

				expectedScript := formatScript("create-env", stateDir, expectedArgs)
				scriptPath := fmt.Sprintf("%s/create-jumpbox.sh", stateDir)
				shellScript, err := fs.ReadFile(scriptPath)
				Expect(err).NotTo(HaveOccurred())

				fileinfo, err := fs.Stat(scriptPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(fileinfo.Mode().String()).To(Equal("-rwxr-x---"))
				Expect(string(shellScript)).To(Equal(expectedScript))

				expectedScript = formatScript("delete-env", stateDir, expectedArgs)
				scriptPath = fmt.Sprintf("%s/delete-jumpbox.sh", stateDir)
				shellScript, err = fs.ReadFile(scriptPath)
				Expect(err).NotTo(HaveOccurred())

				fileinfo, err = fs.Stat(scriptPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(fileinfo.Mode().String()).To(Equal("-rwxr-x---"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(shellScript)).To(Equal(expectedScript))
			})
		})

		Context("on azure", func() {
			It("generates create-env args for jumpbox", func() {
				err := executor.PlanJumpbox(dirInput, deploymentDir, "azure")
				Expect(err).NotTo(HaveOccurred())

				expectedArgs := []string{
					fmt.Sprintf("%s/jumpbox.yml", relativeDeploymentDir),
					"--state", fmt.Sprintf("%s/jumpbox-state.json", relativeVarsDir),
					"--vars-store", fmt.Sprintf("%s/jumpbox-vars-store.yml", relativeVarsDir),
					"--vars-file", fmt.Sprintf("%s/jumpbox-vars-file.yml", relativeVarsDir),
					"-o", fmt.Sprintf("%s/azure/cpi.yml", relativeDeploymentDir),
					"-v", `subscription_id="${BBL_AZURE_SUBSCRIPTION_ID}"`,
					"-v", `client_id="${BBL_AZURE_CLIENT_ID}"`,
					"-v", `client_secret="${BBL_AZURE_CLIENT_SECRET}"`,
					"-v", `tenant_id="${BBL_AZURE_TENANT_ID}"`,
				}

				By("creating the azure cpi file", func() {
					simplePath := filepath.Join(deploymentDir, "azure", "cpi.yml")

					contents, err := fs.ReadFile(simplePath)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(contents)).To(Equal("azure-cpi"))
				})

				By("writing the create-env args to a shell script", func() {
					expectedScript := formatScript("create-env", stateDir, expectedArgs)
					shellScript, err := fs.ReadFile(fmt.Sprintf("%s/create-jumpbox.sh", stateDir))
					Expect(err).NotTo(HaveOccurred())

					Expect(string(shellScript)).To(Equal(expectedScript))
				})

				By("writing the delete-env args to a shell script", func() {
					expectedScript := formatScript("delete-env", stateDir, expectedArgs)
					shellScript, err := fs.ReadFile(fmt.Sprintf("%s/delete-jumpbox.sh", stateDir))
					Expect(err).NotTo(HaveOccurred())

					Expect(string(shellScript)).To(Equal(expectedScript))
				})
			})
		})

		Context("on gcp", func() {
			It("generates create-env args for jumpbox", func() {
				err := executor.PlanJumpbox(dirInput, deploymentDir, "gcp")
				Expect(err).NotTo(HaveOccurred())

				expectedArgs := []string{
					fmt.Sprintf("%s/jumpbox.yml", relativeDeploymentDir),
					"--state", fmt.Sprintf("%s/jumpbox-state.json", relativeVarsDir),
					"--vars-store", fmt.Sprintf("%s/jumpbox-vars-store.yml", relativeVarsDir),
					"--vars-file", fmt.Sprintf("%s/jumpbox-vars-file.yml", relativeVarsDir),
					"-o", fmt.Sprintf("%s/gcp/cpi.yml", relativeDeploymentDir),
					"--var-file", `gcp_credentials_json="${BBL_GCP_SERVICE_ACCOUNT_KEY_PATH}"`,
					"-v", `project_id="${BBL_GCP_PROJECT_ID}"`,
					"-v", `zone="${BBL_GCP_ZONE}"`,
				}

				By("writing the create-env args to a shell script", func() {
					expectedScript := formatScript("create-env", stateDir, expectedArgs)
					shellScript, err := fs.ReadFile(fmt.Sprintf("%s/create-jumpbox.sh", stateDir))
					Expect(err).NotTo(HaveOccurred())

					Expect(string(shellScript)).To(Equal(expectedScript))
				})

				By("writing the delete-env args to a shell script", func() {
					expectedScript := formatScript("delete-env", stateDir, expectedArgs)
					shellScript, err := fs.ReadFile(fmt.Sprintf("%s/delete-jumpbox.sh", stateDir))
					Expect(err).NotTo(HaveOccurred())

					Expect(string(shellScript)).To(Equal(expectedScript))
				})
			})
		})

		Context("when the iaas is vsphere", func() {
			It("generates create-env args for jumpbox", func() {
				err := executor.PlanJumpbox(dirInput, deploymentDir, "vsphere")
				Expect(err).NotTo(HaveOccurred())

				expectedArgs := []string{
					fmt.Sprintf("%s/jumpbox.yml", relativeDeploymentDir),
					"--state", fmt.Sprintf("%s/jumpbox-state.json", relativeVarsDir),
					"--vars-store", fmt.Sprintf("%s/jumpbox-vars-store.yml", relativeVarsDir),
					"--vars-file", fmt.Sprintf("%s/jumpbox-vars-file.yml", relativeVarsDir),
					"-o", fmt.Sprintf("%s/vsphere/cpi.yml", relativeDeploymentDir),
					"-o", fmt.Sprintf("%s/vsphere/resource-pool.yml", relativeDeploymentDir),
					"-o", fmt.Sprintf("%s/vsphere-jumpbox-network.yml", relativeDeploymentDir),
					"-v", `vcenter_user="${BBL_VSPHERE_VCENTER_USER}"`,
					"-v", `vcenter_password="${BBL_VSPHERE_VCENTER_PASSWORD}"`,
				}

				By("writing the jumpbox-network ops-file", func() {
					opsfile, err := fs.ReadFile(fmt.Sprintf("%s/vsphere-jumpbox-network.yml", deploymentDir))
					Expect(err).NotTo(HaveOccurred())

					Expect(string(opsfile)).To(ContainSubstring("instance_groups/name=jumpbox/networks/name=public"))
				})

				By("writing the create-env args to a shell script", func() {
					expectedScript := formatScript("create-env", stateDir, expectedArgs)
					shellScript, err := fs.ReadFile(fmt.Sprintf("%s/create-jumpbox.sh", stateDir))
					Expect(err).NotTo(HaveOccurred())

					Expect(string(shellScript)).To(Equal(expectedScript))
				})

				By("writing the delete-env args to a shell script", func() {
					expectedScript := formatScript("delete-env", stateDir, expectedArgs)
					shellScript, err := fs.ReadFile(fmt.Sprintf("%s/delete-jumpbox.sh", stateDir))
					Expect(err).NotTo(HaveOccurred())

					Expect(string(shellScript)).To(Equal(expectedScript))
				})
			})
		})

		Context("openstack", func() {
			It("generates create-env args for jumpbox", func() {
				err := executor.PlanJumpbox(dirInput, deploymentDir, "openstack")
				Expect(err).NotTo(HaveOccurred())

				expectedArgs := []string{
					fmt.Sprintf("%s/jumpbox.yml", relativeDeploymentDir),
					"--state", fmt.Sprintf("%s/jumpbox-state.json", relativeVarsDir),
					"--vars-store", fmt.Sprintf("%s/jumpbox-vars-store.yml", relativeVarsDir),
					"--vars-file", fmt.Sprintf("%s/jumpbox-vars-file.yml", relativeVarsDir),
					"-o", fmt.Sprintf("%s/openstack/cpi.yml", relativeDeploymentDir),
					"-v", `openstack_username="${BBL_OPENSTACK_USERNAME}"`,
					"-v", `openstack_password="${BBL_OPENSTACK_PASSWORD}"`,
				}

				By("writing the create-env args to a shell script", func() {
					expectedScript := formatScript("create-env", stateDir, expectedArgs)
					shellScript, err := fs.ReadFile(fmt.Sprintf("%s/create-jumpbox.sh", stateDir))
					Expect(err).NotTo(HaveOccurred())

					Expect(string(shellScript)).To(Equal(expectedScript))
				})

				By("writing the delete-env args to a shell script", func() {
					expectedScript := formatScript("delete-env", stateDir, expectedArgs)
					shellScript, err := fs.ReadFile(fmt.Sprintf("%s/delete-jumpbox.sh", stateDir))
					Expect(err).NotTo(HaveOccurred())

					Expect(string(shellScript)).To(Equal(expectedScript))
				})
			})
		})

		Context("cloudstack", func() {
			It("generates create-env args for jumpbox", func() {
				err := executor.PlanJumpbox(dirInput, deploymentDir, "cloudstack")
				Expect(err).NotTo(HaveOccurred())

				expectedArgs := []string{
					fmt.Sprintf("%s/jumpbox.yml", relativeDeploymentDir),
					"--state", fmt.Sprintf("%s/jumpbox-state.json", relativeVarsDir),
					"--vars-store", fmt.Sprintf("%s/jumpbox-vars-store.yml", relativeVarsDir),
					"--vars-file", fmt.Sprintf("%s/jumpbox-vars-file.yml", relativeVarsDir),
					"-o", fmt.Sprintf("%s/cloudstack/cpi.yml", relativeDeploymentDir),
					"-v", `cloudstack_api_key="${BBL_CLOUDSTACK_API_KEY}"`,
					"-v", `cloudstack_secret_access_key="${BBL_CLOUDSTACK_SECRET_ACCESS_KEY}"`,
				}

				By("writing the create-env args to a shell script", func() {
					expectedScript := formatScript("create-env", stateDir, expectedArgs)
					shellScript, err := fs.ReadFile(fmt.Sprintf("%s/create-jumpbox.sh", stateDir))
					Expect(err).NotTo(HaveOccurred())

					Expect(string(shellScript)).To(Equal(expectedScript))
				})

				By("writing the delete-env args to a shell script", func() {
					expectedScript := formatScript("delete-env", stateDir, expectedArgs)
					shellScript, err := fs.ReadFile(fmt.Sprintf("%s/delete-jumpbox.sh", stateDir))
					Expect(err).NotTo(HaveOccurred())

					Expect(string(shellScript)).To(Equal(expectedScript))
				})
			})
		})
	})

	Describe("PlanDirector", func() {
		It("writes bosh-deployment assets to the deployment dir", func() {
			err := executor.PlanDirector(dirInput, deploymentDir, "warden")
			Expect(err).NotTo(HaveOccurred())

			simplePath := filepath.Join(deploymentDir, "LICENSE")
			contents, err := fs.ReadFile(simplePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("my-license"))

			nestedPath := filepath.Join(deploymentDir, "vsphere", "cpi.yml")
			contents, err = fs.ReadFile(nestedPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("vsphere-cpi"))
		})

		Context("aws", func() {
			It("writes create-director.sh and delete-director.sh", func() {
				expectedArgs := []string{
					filepath.Join(relativeDeploymentDir, "bosh.yml"),
					"--state", filepath.Join(relativeVarsDir, "bosh-state.json"),
					"--vars-store", filepath.Join(relativeVarsDir, "director-vars-store.yml"),
					"--vars-file", filepath.Join(relativeVarsDir, "director-vars-file.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "aws", "cpi.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "jumpbox-user.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "uaa.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "credhub.yml"),
					"-o", filepath.Join(relativeStateDir, "bbl-ops-files", "aws", "bosh-director-ephemeral-ip-ops.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "aws", "iam-instance-profile.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "aws", "encrypted-disk.yml"),
					"-v", `access_key_id="${BBL_AWS_ACCESS_KEY_ID}"`,
					"-v", `secret_access_key="${BBL_AWS_SECRET_ACCESS_KEY}"`,
				}

				behavesLikePlan(expectedArgs, cli, fs, executor, dirInput, deploymentDir, "aws", stateDir)
			})

			It("writes aws-specific ops files", func() {
				err := executor.PlanDirector(dirInput, deploymentDir, "aws")
				Expect(err).NotTo(HaveOccurred())

				ipOpsFile := filepath.Join(stateDir, "bbl-ops-files", "aws", "bosh-director-ephemeral-ip-ops.yml")

				ipOpsFileContents, err := fs.ReadFile(ipOpsFile)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(ipOpsFileContents)).To(Equal(`
- type: replace
  path: /resource_pools/name=vms/cloud_properties/auto_assign_public_ip?
  value: true
`))
			})
		})

		Context("gcp", func() {
			It("writes create-director.sh and delete-director.sh", func() {
				expectedArgs := []string{
					filepath.Join(relativeDeploymentDir, "bosh.yml"),
					"--state", filepath.Join(relativeVarsDir, "bosh-state.json"),
					"--vars-store", filepath.Join(relativeVarsDir, "director-vars-store.yml"),
					"--vars-file", filepath.Join(relativeVarsDir, "director-vars-file.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "gcp", "cpi.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "jumpbox-user.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "uaa.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "credhub.yml"),
					"-o", filepath.Join(relativeStateDir, "bbl-ops-files", "gcp", "bosh-director-ephemeral-ip-ops.yml"),
					"--var-file", `gcp_credentials_json="${BBL_GCP_SERVICE_ACCOUNT_KEY_PATH}"`,
					"-v", `project_id="${BBL_GCP_PROJECT_ID}"`,
					"-v", `zone="${BBL_GCP_ZONE}"`,
				}

				behavesLikePlan(expectedArgs, cli, fs, executor, dirInput, deploymentDir, "gcp", stateDir)
			})

			It("writes gcp-specific ops files", func() {
				err := executor.PlanDirector(dirInput, deploymentDir, "gcp")
				Expect(err).NotTo(HaveOccurred())

				ipOpsFile := filepath.Join(stateDir, "bbl-ops-files", "gcp", "bosh-director-ephemeral-ip-ops.yml")

				ipOpsFileContents, err := fs.ReadFile(ipOpsFile)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(ipOpsFileContents)).To(Equal(`
- type: replace
  path: /networks/name=default/subnets/0/cloud_properties/ephemeral_external_ip?
  value: true
`))
			})
		})

		Context("azure", func() {
			It("writes create-director.sh and delete-director.sh", func() {
				expectedArgs := []string{
					filepath.Join(relativeDeploymentDir, "bosh.yml"),
					"--state", filepath.Join(relativeVarsDir, "bosh-state.json"),
					"--vars-store", filepath.Join(relativeVarsDir, "director-vars-store.yml"),
					"--vars-file", filepath.Join(relativeVarsDir, "director-vars-file.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "azure", "cpi.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "jumpbox-user.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "uaa.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "credhub.yml"),
					"-v", `subscription_id="${BBL_AZURE_SUBSCRIPTION_ID}"`,
					"-v", `client_id="${BBL_AZURE_CLIENT_ID}"`,
					"-v", `client_secret="${BBL_AZURE_CLIENT_SECRET}"`,
					"-v", `tenant_id="${BBL_AZURE_TENANT_ID}"`,
				}

				behavesLikePlan(expectedArgs, cli, fs, executor, dirInput, deploymentDir, "azure", stateDir)
			})
		})

		Context("vsphere", func() {
			It("writes create-director.sh and delete-director.sh", func() {
				expectedArgs := []string{
					filepath.Join(relativeDeploymentDir, "bosh.yml"),
					"--state", filepath.Join(relativeVarsDir, "bosh-state.json"),
					"--vars-store", filepath.Join(relativeVarsDir, "director-vars-store.yml"),
					"--vars-file", filepath.Join(relativeVarsDir, "director-vars-file.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "vsphere", "cpi.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "jumpbox-user.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "uaa.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "credhub.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "vsphere", "resource-pool.yml"),
					"-v", `vcenter_user="${BBL_VSPHERE_VCENTER_USER}"`,
					"-v", `vcenter_password="${BBL_VSPHERE_VCENTER_PASSWORD}"`,
				}

				behavesLikePlan(expectedArgs, cli, fs, executor, dirInput, deploymentDir, "vsphere", stateDir)
			})
		})

		Context("openstack", func() {
			It("writes create-director.sh and delete-director.sh", func() {
				expectedArgs := []string{
					filepath.Join(relativeDeploymentDir, "bosh.yml"),
					"--state", filepath.Join(relativeVarsDir, "bosh-state.json"),
					"--vars-store", filepath.Join(relativeVarsDir, "director-vars-store.yml"),
					"--vars-file", filepath.Join(relativeVarsDir, "director-vars-file.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "openstack", "cpi.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "jumpbox-user.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "uaa.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "credhub.yml"),
					"-v", `openstack_username="${BBL_OPENSTACK_USERNAME}"`,
					"-v", `openstack_password="${BBL_OPENSTACK_PASSWORD}"`,
				}

				behavesLikePlan(expectedArgs, cli, fs, executor, dirInput, deploymentDir, "openstack", stateDir)
			})
		})
		Context("cloudstack", func() {
			It("writes create-director.sh and delete-director.sh", func() {
				expectedArgs := []string{
					filepath.Join(relativeDeploymentDir, "bosh.yml"),
					"--state", filepath.Join(relativeVarsDir, "bosh-state.json"),
					"--vars-store", filepath.Join(relativeVarsDir, "director-vars-store.yml"),
					"--vars-file", filepath.Join(relativeVarsDir, "director-vars-file.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "cloudstack", "cpi.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "jumpbox-user.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "uaa.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "credhub.yml"),
					"-v", `cloudstack_api_key="${BBL_CLOUDSTACK_API_KEY}"`,
					"-v", `cloudstack_secret_access_key="${BBL_CLOUDSTACK_SECRET_ACCESS_KEY}"`,
				}

				behavesLikePlan(expectedArgs, cli, fs, executor, dirInput, deploymentDir, "cloudstack", stateDir)
			})
		})
	})

	Describe("WriteDeploymentVars", func() {
		BeforeEach(func() {
			dirInput.Deployment = "some-deployment"
		})

		It("writes the deployment vars yml file", func() {
			err := executor.WriteDeploymentVars(dirInput, "some-deployment-vars")
			Expect(err).NotTo(HaveOccurred())

			deploymentVars, err := fs.ReadFile(filepath.Join(varsDir, "some-deployment-vars-file.yml"))
			Expect(err).NotTo(HaveOccurred())

			Expect(string(deploymentVars)).To(Equal("some-deployment-vars"))
		})
	})

	Describe("CreateEnv", func() {
		var (
			cli      *fakes.BOSHCLI
			executor bosh.Executor

			createEnvPath string
			varsDir       string
			stateDir      string

			dirInput bosh.DirInput
			state    storage.State
		)

		BeforeEach(func() {
			fs = &afero.Afero{Fs: afero.NewOsFs()} // real os fs so we can exec scripts...
			cli = &fakes.BOSHCLI{}

			var err error
			varsDir, err = fs.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())
			stateDir, err = fs.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			box := packr.New("another-test", "./deployments")

			executor = bosh.Executor{
				CLI: cli,
				FS:  fs,
				Box: box,
			}

			dirInput = bosh.DirInput{
				Deployment: "some-deployment",
				StateDir:   stateDir,
				VarsDir:    varsDir,
			}

			createEnvPath = filepath.Join(stateDir, "create-some-deployment.sh")
			createEnvContents := fmt.Sprintf("#!/bin/bash\necho 'some-vars-store-contents' > %s/some-deployment-vars-store.yml\n", varsDir)

			fs.WriteFile(createEnvPath, []byte(createEnvContents), storage.ScriptMode)
		})

		AfterEach(func() {
			fs.Remove(filepath.Join(varsDir, "some-deployment-vars-store.yml"))
			fs.Remove(createEnvPath)
			fs.Remove(filepath.Join(stateDir, "create-some-deployment-override.sh"))
			os.Unsetenv("BBL_STATE_DIR")
		})

		Context("when the user provides a create-env override", func() {
			BeforeEach(func() {
				overridePath := filepath.Join(stateDir, "create-some-deployment-override.sh")
				overrideContents := fmt.Sprintf("#!/bin/bash\necho 'override-vars-store-contents' > %s/some-deployment-vars-store.yml\n", varsDir)

				fs.WriteFile(overridePath, []byte(overrideContents), storage.ScriptMode)
			})

			It("runs the create-env-override.sh script", func() {
				vars, err := executor.CreateEnv(dirInput, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(cli.RunCallCount()).To(Equal(0))
				Expect(vars).To(ContainSubstring("override-vars-store-contents"))
			})
		})

		Context("when the user provides a create-env override with flag-provided environment variable", func() {
			BeforeEach(func() {
				err := os.Unsetenv("BBL_IAAS")
				Expect(err).NotTo(HaveOccurred())

				state.IAAS = "some-fictional-iaas"
				overrideContents := fmt.Sprintf("#!/bin/bash\n [ \"${BBL_IAAS}\" = \"some-fictional-iaas\" ]")

				overridePath := filepath.Join(stateDir, "create-some-deployment-override.sh")
				fs.WriteFile(overridePath, []byte(overrideContents), storage.ScriptMode)
			})

			It("runs the create-env-override.sh script", func() {
				_, err := executor.CreateEnv(dirInput, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(cli.RunCallCount()).To(Equal(0))
			})
		})

		It("runs the create-env script and returns the resulting vars-store contents", func() {
			vars, err := executor.CreateEnv(dirInput, state)
			Expect(err).NotTo(HaveOccurred())

			Expect(cli.RunCallCount()).To(Equal(0))
			Expect(vars).To(ContainSubstring("some-vars-store-contents"))

			By("setting BBL_STATE_DIR environment variable", func() {
				bblStateDirEnv := os.Getenv("BBL_STATE_DIR")
				Expect(bblStateDirEnv).To(Equal(stateDir))
			})
		})

		Context("when iaas credentials are provided", func() {
			Context("on aws", func() {
				BeforeEach(func() {
					state.IAAS = "aws"
					state.AWS = storage.AWS{
						AccessKeyID:     "some-access-key-id",
						SecretAccessKey: "some-secret-access-key",
					}
				})

				It("sets credentials in environment variables", func() {
					_, err := executor.CreateEnv(dirInput, state)
					Expect(err).NotTo(HaveOccurred())

					Expect(os.Getenv("BBL_AWS_ACCESS_KEY_ID")).To(Equal("some-access-key-id"))
					Expect(os.Getenv("BBL_AWS_SECRET_ACCESS_KEY")).To(Equal("some-secret-access-key"))
				})
			})

			Context("on azure", func() {
				BeforeEach(func() {
					state.IAAS = "azure"
					state.Azure = storage.Azure{
						ClientID:       "some-client-id",
						ClientSecret:   "some-client-secret",
						SubscriptionID: "some-subscription-id",
						TenantID:       "some-tenant-id",
					}
				})

				It("sets credentials in environment variables", func() {
					_, err := executor.CreateEnv(dirInput, state)
					Expect(err).NotTo(HaveOccurred())

					Expect(os.Getenv("BBL_AZURE_CLIENT_ID")).To(Equal("some-client-id"))
					Expect(os.Getenv("BBL_AZURE_CLIENT_SECRET")).To(Equal("some-client-secret"))
					Expect(os.Getenv("BBL_AZURE_SUBSCRIPTION_ID")).To(Equal("some-subscription-id"))
					Expect(os.Getenv("BBL_AZURE_TENANT_ID")).To(Equal("some-tenant-id"))
				})
			})

			Context("on gcp", func() {
				BeforeEach(func() {
					state.IAAS = "gcp"
					state.GCP = storage.GCP{
						ServiceAccountKeyPath: "some-service-account-key-path",
						Zone:                  "some-zone",
						ProjectID:             "some-project-id",
					}
				})

				It("sets credentials in environment variables", func() {
					_, err := executor.CreateEnv(dirInput, state)
					Expect(err).NotTo(HaveOccurred())

					Expect(os.Getenv("BBL_GCP_SERVICE_ACCOUNT_KEY_PATH")).To(Equal("some-service-account-key-path"))
					Expect(os.Getenv("BBL_GCP_ZONE")).To(Equal("some-zone"))
					Expect(os.Getenv("BBL_GCP_PROJECT_ID")).To(Equal("some-project-id"))
				})
			})

			Context("on vsphere", func() {
				It("sets credentials in environment variables", func() {
					_, err := executor.CreateEnv(dirInput, storage.State{
						IAAS: "vsphere",
						VSphere: storage.VSphere{
							VCenterUser:     "some-user",
							VCenterPassword: "some-password",
						},
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(os.Getenv("BBL_VSPHERE_VCENTER_USER")).To(Equal("some-user"))
					Expect(os.Getenv("BBL_VSPHERE_VCENTER_PASSWORD")).To(Equal("some-password"))
				})
			})

			Context("on openstack", func() {
				It("sets credentials in environment variables", func() {
					_, err := executor.CreateEnv(dirInput, storage.State{
						IAAS: "openstack",
						OpenStack: storage.OpenStack{
							Username: "some-user",
							Password: "some-password",
						},
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(os.Getenv("BBL_OPENSTACK_USERNAME")).To(Equal("some-user"))
					Expect(os.Getenv("BBL_OPENSTACK_PASSWORD")).To(Equal("some-password"))
				})
			})
			Context("on cloudstack", func() {
				It("sets credentials in environment variables", func() {
					_, err := executor.CreateEnv(dirInput, storage.State{
						IAAS: "cloudstack",
						CloudStack: storage.CloudStack{
							ApiKey:          "some-api-key",
							SecretAccessKey: "some-secret",
						},
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(os.Getenv("BBL_CLOUDSTACK_API_KEY")).To(Equal("some-api-key"))
					Expect(os.Getenv("BBL_CLOUDSTACK_SECRET_ACCESS_KEY")).To(Equal("some-secret"))
				})
			})
		})

		Context("when the create-env script returns an error", func() {
			BeforeEach(func() {
				createEnvContents := "#!/bin/bash\nexit 1\n"
				fs.WriteFile(createEnvPath, []byte(createEnvContents), storage.ScriptMode)
			})

			It("returns an error", func() {
				vars, err := executor.CreateEnv(dirInput, state)
				Expect(err).To(MatchError(fmt.Sprintf("Running %s: exit status 1", createEnvPath)))
				Expect(vars).To(Equal(""))
			})
		})
	})

	Describe("DeleteEnv", func() {
		var (
			executor bosh.Executor

			deleteEnvPath string
			varsDir       string
			stateDir      string

			dirInput bosh.DirInput
			state    storage.State
		)

		BeforeEach(func() {
			fs = &afero.Afero{Fs: afero.NewOsFs()} // real os fs so we can exec scripts...

			var err error
			varsDir, err = fs.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())
			stateDir, err = fs.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			box := packr.New("best-test-box", "./deployments")

			executor = bosh.Executor{
				CLI: cli,
				FS:  fs,
				Box: box,
			}

			dirInput = bosh.DirInput{
				Deployment: "director",
				VarsDir:    varsDir,
				StateDir:   stateDir,
			}

			state = storage.State{
				IAAS: "some-iaas",
			}

			deleteEnvPath = filepath.Join(stateDir, "delete-director.sh")
			deleteEnvContents := "#!/bin/bash\necho delete-env > /dev/null\n"
			fs.WriteFile(deleteEnvPath, []byte(deleteEnvContents), storage.ScriptMode)

			deploymentStateJson := filepath.Join(varsDir, "bosh-state.json")
			fs.WriteFile(deploymentStateJson, []byte("some: deployment"), storage.StateMode)
		})

		AfterEach(func() {
			os.Unsetenv("BBL_STATE_DIR")
			fs.Remove(filepath.Join(stateDir, "delete-director.sh"))
		})

		Context("when the user provides a delete-env override", func() {
			BeforeEach(func() {
				overridePath := filepath.Join(stateDir, "delete-director-override.sh")
				overrideContents := fmt.Sprintf("#!/bin/bash\necho 'override' > %s/delete-env-output\n", varsDir)

				fs.WriteFile(overridePath, []byte(overrideContents), storage.ScriptMode)
			})

			AfterEach(func() {
				fs.Remove(filepath.Join(varsDir, "delete-env-output"))
				fs.Remove(filepath.Join(stateDir, "delete-director-override.sh"))
			})

			It("runs the delete-env-override.sh script", func() {
				err := executor.DeleteEnv(dirInput, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(cli.RunCallCount()).To(Equal(0))

				overrideOut, err := fs.ReadFile(filepath.Join(varsDir, "delete-env-output"))
				Expect(err).NotTo(HaveOccurred())
				Expect(overrideOut).To(ContainSubstring("override"))
			})
		})

		Context("when the user tries to delete a jumpbox", func() {
			BeforeEach(func() {
				dirInput.Deployment = "jumpbox"
				deleteEnvPath = filepath.Join(stateDir, "delete-jumpbox.sh")
				deleteEnvContents := "#!/bin/bash\necho delete-env > /dev/null\n"
				fs.WriteFile(deleteEnvPath, []byte(deleteEnvContents), storage.ScriptMode)

				deploymentStateJson := filepath.Join(varsDir, "jumpbox-state.json")
				fs.WriteFile(deploymentStateJson, []byte("some: deployment"), storage.StateMode)
			})

			AfterEach(func() {
				fs.Remove(filepath.Join(stateDir, "delete-jumpbox.sh"))
				fs.Remove(filepath.Join(stateDir, "jumpbox-state.json"))
			})

			It("deletes a bosh environment with the delete-env script", func() {
				err := executor.DeleteEnv(dirInput, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(cli.RunCallCount()).To(Equal(0))

				By("setting BBL_STATE_DIR environment variable", func() {
					bblStateDirEnv := os.Getenv("BBL_STATE_DIR")
					Expect(bblStateDirEnv).To(Equal(stateDir))
				})
			})
		})

		Context("when the user tries to delete an unfamiliar deployment-type-thing", func() {
			BeforeEach(func() {
				dirInput.Deployment = "garbaggio-deployment"
			})

			It("errors reasonably", func() {
				err := executor.DeleteEnv(dirInput, state)
				Expect(err).To(HaveOccurred())
			})
		})

		It("deletes a bosh environment with the delete-env script", func() {
			err := executor.DeleteEnv(dirInput, state)
			Expect(err).NotTo(HaveOccurred())

			Expect(cli.RunCallCount()).To(Equal(0))

			By("setting BBL_STATE_DIR environment variable", func() {
				bblStateDirEnv := os.Getenv("BBL_STATE_DIR")
				Expect(bblStateDirEnv).To(Equal(stateDir))
			})
		})

		Context("when iaas credentials are provided", func() {
			Context("on aws", func() {
				BeforeEach(func() {
					state.IAAS = "aws"
					state.AWS = storage.AWS{
						AccessKeyID:     "some-access-key-id",
						SecretAccessKey: "some-secret-access-key",
					}
				})

				It("sets credentials in environment variables", func() {
					err := executor.DeleteEnv(dirInput, state)
					Expect(err).NotTo(HaveOccurred())

					Expect(os.Getenv("BBL_AWS_ACCESS_KEY_ID")).To(Equal("some-access-key-id"))
					Expect(os.Getenv("BBL_AWS_SECRET_ACCESS_KEY")).To(Equal("some-secret-access-key"))
				})
			})

			Context("on azure", func() {
				BeforeEach(func() {
					state.IAAS = "azure"
					state.Azure = storage.Azure{
						ClientID:       "some-client-id",
						ClientSecret:   "some-client-secret",
						SubscriptionID: "some-subscription-id",
						TenantID:       "some-tenant-id",
					}
				})

				It("sets credentials in environment variables", func() {
					err := executor.DeleteEnv(dirInput, state)
					Expect(err).NotTo(HaveOccurred())

					Expect(os.Getenv("BBL_AZURE_CLIENT_ID")).To(Equal("some-client-id"))
					Expect(os.Getenv("BBL_AZURE_CLIENT_SECRET")).To(Equal("some-client-secret"))
					Expect(os.Getenv("BBL_AZURE_SUBSCRIPTION_ID")).To(Equal("some-subscription-id"))
					Expect(os.Getenv("BBL_AZURE_TENANT_ID")).To(Equal("some-tenant-id"))
				})
			})

			Context("on gcp", func() {
				BeforeEach(func() {
					state.IAAS = "gcp"
					state.GCP = storage.GCP{
						ServiceAccountKeyPath: "some-service-account-key-path",
						Zone:                  "some-zone",
						ProjectID:             "some-project-id",
					}
				})

				It("sets credentials in environment variables", func() {
					err := executor.DeleteEnv(dirInput, state)
					Expect(err).NotTo(HaveOccurred())

					Expect(os.Getenv("BBL_GCP_SERVICE_ACCOUNT_KEY_PATH")).To(Equal("some-service-account-key-path"))
					Expect(os.Getenv("BBL_GCP_ZONE")).To(Equal("some-zone"))
					Expect(os.Getenv("BBL_GCP_PROJECT_ID")).To(Equal("some-project-id"))
				})
			})

			Context("on vsphere", func() {
				BeforeEach(func() {
					state.IAAS = "vsphere"
					state.VSphere = storage.VSphere{
						VCenterUser:     "some-user",
						VCenterPassword: "some-password",
					}
				})

				It("sets credentials in environment variables", func() {
					err := executor.DeleteEnv(dirInput, state)
					Expect(err).NotTo(HaveOccurred())

					Expect(os.Getenv("BBL_VSPHERE_VCENTER_USER")).To(Equal("some-user"))
					Expect(os.Getenv("BBL_VSPHERE_VCENTER_PASSWORD")).To(Equal("some-password"))
				})
			})
			Context("on cloudstack", func() {
				BeforeEach(func() {
					state.IAAS = "cloudstack"
					state.CloudStack = storage.CloudStack{
						ApiKey:          "some-api-key",
						SecretAccessKey: "some-secret",
					}
				})

				It("sets credentials in environment variables", func() {
					err := executor.DeleteEnv(dirInput, state)
					Expect(err).NotTo(HaveOccurred())

					Expect(os.Getenv("BBL_CLOUDSTACK_API_KEY")).To(Equal("some-api-key"))
					Expect(os.Getenv("BBL_CLOUDSTACK_SECRET_ACCESS_KEY")).To(Equal("some-secret"))
				})
			})
		})

		Context("when the create-env script returns an error", func() {
			BeforeEach(func() {
				deleteEnvContents := "#!/bin/bash\nexit 1\n"
				fs.WriteFile(deleteEnvPath, []byte(deleteEnvContents), storage.ScriptMode)
			})

			It("returns an error", func() {
				err := executor.DeleteEnv(dirInput, state)
				Expect(err).To(MatchError("Run bosh delete-env director: exit status 1"))
			})
		})
	})

	Describe("Version", func() {
		var (
			cli      *fakes.BOSHCLI
			executor bosh.Executor
		)
		BeforeEach(func() {
			cli = &fakes.BOSHCLI{}
			cli.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
				stdout.Write([]byte("some-text version 1.1.1 some-other-text"))
				return nil
			}

			executor = bosh.NewExecutor(cli, fs)
		})

		It("returns the correctly trimmed version", func() {
			version, err := executor.Version()
			Expect(err).NotTo(HaveOccurred())

			_, _, args := cli.RunArgsForCall(0)
			Expect(args).To(Equal([]string{"-v"}))

			Expect(version).To(Equal("1.1.1"))
		})

		Context("when the run cli fails", func() {
			BeforeEach(func() {
				cli.RunStub = nil
				cli.RunCall.Returns.Error = errors.New("banana")
			})

			It("returns an error", func() {
				_, err := executor.Version()
				Expect(err).To(MatchError("banana"))
			})
		})

		Context("when the version cannot be parsed", func() {
			BeforeEach(func() {
				cli.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
					stdout.Write([]byte(""))
					return nil
				}
			})

			It("returns a bosh version error", func() {
				_, err := executor.Version()
				Expect(err).To(MatchError("BOSH version could not be parsed"))
			})
		})
	})
})

func formatScript(command string, stateDir string, args []string) string {
	script := fmt.Sprintf("#!/bin/sh\nbosh-path %s \\\n", command)
	for _, arg := range args {
		if arg[0] == '-' {
			script = fmt.Sprintf("%s  %s", script, arg)
		} else {
			script = fmt.Sprintf("%s  %s \\\n", script, arg)
		}
	}

	return fmt.Sprintf("%s\n", script[:len(script)-2])
}

type behavesLikePlanFs interface {
	fileio.FileReader
	fileio.Stater
}

func behavesLikePlan(expectedArgs []string, cli *fakes.BOSHCLI, fs behavesLikePlanFs, executor bosh.Executor, input bosh.DirInput, deploymentDir, iaas, stateDir string) {
	cli.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
		stdout.Write([]byte("some-manifest"))
		return nil
	}

	err := executor.PlanDirector(input, deploymentDir, iaas)
	Expect(err).NotTo(HaveOccurred())
	Expect(cli.RunCallCount()).To(Equal(0))

	By("writing the create-env args to a shell script", func() {
		expectedScript := formatScript("create-env", stateDir, expectedArgs)
		scriptPath := fmt.Sprintf("%s/create-director.sh", stateDir)
		shellScript, err := fs.ReadFile(scriptPath)
		Expect(err).NotTo(HaveOccurred())

		fileinfo, err := fs.Stat(scriptPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(fileinfo.Mode().String()).To(Equal("-rwxr-x---"))
		Expect(err).NotTo(HaveOccurred())

		Expect(string(shellScript)).To(Equal(expectedScript))
	})

	By("writing the delete-env args to a shell script", func() {
		expectedScript := formatScript("delete-env", stateDir, expectedArgs)
		scriptPath := fmt.Sprintf("%s/delete-director.sh", stateDir)
		shellScript, err := fs.ReadFile(scriptPath)
		Expect(err).NotTo(HaveOccurred())

		fileinfo, err := fs.Stat(scriptPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(fileinfo.Mode().String()).To(Equal("-rwxr-x---"))
		Expect(err).NotTo(HaveOccurred())
		Expect(err).NotTo(HaveOccurred())

		Expect(string(shellScript)).To(Equal(expectedScript))
	})
}
