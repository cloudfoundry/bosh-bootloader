package bosh_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Executor", func() {
	Describe("PlanJumpbox", func() {
		var (
			cmd *fakes.BOSHCommand

			stateDir              string
			deploymentDir         string
			relativeDeploymentDir string
			relativeVarsDir       string

			executor bosh.Executor
			dirInput bosh.DirInput
			state    storage.State
		)

		BeforeEach(func() {
			cmd = &fakes.BOSHCommand{}
			cmd.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
				stdout.Write([]byte("some-manifest"))
				return nil
			}
			cmd.GetBOSHPathCall.Returns.Path = "bosh-path"

			var err error
			stateDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			deploymentDir = filepath.Join(stateDir, "deployment")
			err = os.Mkdir(deploymentDir, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			varsDir := filepath.Join(stateDir, "vars")
			err = os.Mkdir(varsDir, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			relativeDeploymentDir = "${BBL_STATE_DIR}/deployment"
			relativeVarsDir = "${BBL_STATE_DIR}/vars"

			dirInput = bosh.DirInput{
				VarsDir:  varsDir,
				StateDir: stateDir,
			}

			state = storage.State{
				IAAS: "aws",
			}

			executor = bosh.NewExecutor(cmd, ioutil.ReadFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)
		})

		It("writes bosh-deployment assets to the deployment dir", func() {
			err := executor.PlanJumpbox(dirInput, deploymentDir, "aws")
			Expect(err).NotTo(HaveOccurred())

			simplePath := filepath.Join(deploymentDir, "no-external-ip.yml")
			expectedContents := bosh.MustAsset("vendor/github.com/cppforlife/jumpbox-deployment/no-external-ip.yml")

			contents, err := ioutil.ReadFile(simplePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).To(Equal(expectedContents))

			nestedPath := filepath.Join(deploymentDir, "vsphere", "cpi.yml")
			expectedContents = bosh.MustAsset("vendor/github.com/cppforlife/jumpbox-deployment/vsphere/cpi.yml")

			contents, err = ioutil.ReadFile(nestedPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).To(Equal(expectedContents))
		})

		It("generates create-env args for jumpbox", func() {
			err := executor.PlanJumpbox(dirInput, deploymentDir, "aws")
			Expect(err).NotTo(HaveOccurred())

			expectedArgs := []string{
				fmt.Sprintf("%s/jumpbox.yml", relativeDeploymentDir),
				"--state", fmt.Sprintf("%s/jumpbox-state.json", relativeVarsDir),
				"--vars-store", fmt.Sprintf("%s/jumpbox-vars-store.yml", relativeVarsDir),
				"--vars-file", fmt.Sprintf("%s/jumpbox-vars-file.yml", relativeVarsDir),
				"-o", fmt.Sprintf("%s/aws/cpi.yml", relativeDeploymentDir),
			}

			By("writing the create-env args to a shell script", func() {
				expectedScript := formatScript("create-env", stateDir, expectedArgs)
				scriptPath := fmt.Sprintf("%s/create-jumpbox.sh", stateDir)
				shellScript, err := ioutil.ReadFile(scriptPath)
				Expect(err).NotTo(HaveOccurred())

				fileinfo, err := os.Stat(scriptPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(fileinfo.Mode().String()).To(Equal("-rwxr-x---"))

				Expect(string(shellScript)).To(Equal(expectedScript))
			})

			By("writing the delete-env args to a shell script", func() {
				expectedScript := formatScript("delete-env", stateDir, expectedArgs)
				scriptPath := fmt.Sprintf("%s/delete-jumpbox.sh", stateDir)
				shellScript, err := ioutil.ReadFile(scriptPath)
				Expect(err).NotTo(HaveOccurred())

				fileinfo, err := os.Stat(scriptPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(fileinfo.Mode().String()).To(Equal("-rwxr-x---"))
				Expect(err).NotTo(HaveOccurred())

				Expect(string(shellScript)).To(Equal(expectedScript))
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
				}

				By("writing the jumpbox-network ops-file", func() {
					opsfile, err := ioutil.ReadFile(fmt.Sprintf("%s/vsphere-jumpbox-network.yml", deploymentDir))
					Expect(err).NotTo(HaveOccurred())

					Expect(string(opsfile)).To(ContainSubstring("instance_groups/name=jumpbox/networks/name=public"))
				})

				By("writing the create-env args to a shell script", func() {
					expectedScript := formatScript("create-env", stateDir, expectedArgs)
					shellScript, err := ioutil.ReadFile(fmt.Sprintf("%s/create-jumpbox.sh", stateDir))
					Expect(err).NotTo(HaveOccurred())

					Expect(string(shellScript)).To(Equal(expectedScript))
				})

				By("writing the delete-env args to a shell script", func() {
					expectedScript := formatScript("delete-env", stateDir, expectedArgs)
					shellScript, err := ioutil.ReadFile(fmt.Sprintf("%s/delete-jumpbox.sh", stateDir))
					Expect(err).NotTo(HaveOccurred())

					Expect(string(shellScript)).To(Equal(expectedScript))
				})
			})
		})
	})

	Describe("PlanDirector", func() {
		var (
			cmd *fakes.BOSHCommand

			stateDir              string
			deploymentDir         string
			relativeDeploymentDir string
			relativeVarsDir       string
			relativeStateDir      string

			executor bosh.Executor
			dirInput bosh.DirInput
			state    storage.State
		)

		BeforeEach(func() {
			cmd = &fakes.BOSHCommand{}
			cmd.GetBOSHPathCall.Returns.Path = "bosh-path"

			var err error
			stateDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			deploymentDir = filepath.Join(stateDir, "deployment")
			err = os.Mkdir(deploymentDir, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			varsDir := filepath.Join(stateDir, "vars")
			err = os.Mkdir(varsDir, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			relativeDeploymentDir = "${BBL_STATE_DIR}/deployment"
			relativeVarsDir = "${BBL_STATE_DIR}/vars"
			relativeStateDir = "${BBL_STATE_DIR}"

			dirInput = bosh.DirInput{
				VarsDir:  varsDir,
				StateDir: stateDir,
			}

			state = storage.State{
				IAAS: "aws",
			}

			executor = bosh.NewExecutor(cmd, ioutil.ReadFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)
		})

		It("writes bosh-deployment assets to the deployment dir", func() {
			err := executor.PlanDirector(dirInput, deploymentDir, "warden")
			Expect(err).NotTo(HaveOccurred())

			simplePath := filepath.Join(deploymentDir, "LICENSE")
			expectedContents := bosh.MustAsset("vendor/github.com/cloudfoundry/bosh-deployment/LICENSE")

			contents, err := ioutil.ReadFile(simplePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).To(Equal(expectedContents))

			nestedPath := filepath.Join(deploymentDir, "vsphere", "cpi.yml")
			expectedContents = bosh.MustAsset("vendor/github.com/cloudfoundry/bosh-deployment/vsphere/cpi.yml")

			contents, err = ioutil.ReadFile(nestedPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).To(Equal(expectedContents))
		})

		Context("aws", func() {
			var expectedArgs []string

			BeforeEach(func() {
				expectedArgs = []string{
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
					"-o", filepath.Join(relativeStateDir, "bbl-ops-files", "aws", "bosh-director-encrypt-disk-ops.yml"),
				}
			})

			It("writes create-director.sh and delete-director.sh", func() {
				behavesLikePlan(expectedArgs, cmd, executor, dirInput, deploymentDir, "aws", stateDir)
			})

			It("writes aws-specific ops files", func() {
				err := executor.PlanDirector(dirInput, deploymentDir, "aws")
				Expect(err).NotTo(HaveOccurred())

				ipOpsFile := filepath.Join(stateDir, "bbl-ops-files", "aws", "bosh-director-ephemeral-ip-ops.yml")
				encryptDiskOpsFile := filepath.Join(stateDir, "bbl-ops-files", "aws", "bosh-director-encrypt-disk-ops.yml")

				ipOpsFileContents, err := ioutil.ReadFile(ipOpsFile)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(ipOpsFileContents)).To(Equal(`
- type: replace
  path: /resource_pools/name=vms/cloud_properties/auto_assign_public_ip?
  value: true
`))
				encryptDiskOpsFileContents, err := ioutil.ReadFile(encryptDiskOpsFile)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(encryptDiskOpsFileContents)).To(Equal(`---
- type: replace
  path: /disk_pools/name=disks/cloud_properties?
  value:
    type: gp2
    encrypted: true
    kms_key_arn: ((kms_key_arn))
`))
			})
		})

		Context("gcp", func() {
			var expectedArgs []string

			BeforeEach(func() {
				expectedArgs = []string{
					filepath.Join(relativeDeploymentDir, "bosh.yml"),
					"--state", filepath.Join(relativeVarsDir, "bosh-state.json"),
					"--vars-store", filepath.Join(relativeVarsDir, "director-vars-store.yml"),
					"--vars-file", filepath.Join(relativeVarsDir, "director-vars-file.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "gcp", "cpi.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "jumpbox-user.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "uaa.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "credhub.yml"),
					"-o", filepath.Join(relativeStateDir, "bbl-ops-files", "gcp", "bosh-director-ephemeral-ip-ops.yml"),
				}
			})

			It("writes create-director.sh and delete-director.sh", func() {
				behavesLikePlan(expectedArgs, cmd, executor, dirInput, deploymentDir, "gcp", stateDir)
			})

			It("writes gcp-specific ops files", func() {
				err := executor.PlanDirector(dirInput, deploymentDir, "gcp")
				Expect(err).NotTo(HaveOccurred())

				ipOpsFile := filepath.Join(stateDir, "bbl-ops-files", "gcp", "bosh-director-ephemeral-ip-ops.yml")

				ipOpsFileContents, err := ioutil.ReadFile(ipOpsFile)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(ipOpsFileContents)).To(Equal(`
- type: replace
  path: /networks/name=default/subnets/0/cloud_properties/ephemeral_external_ip?
  value: true
`))
			})
		})

		Context("azure", func() {
			var expectedArgs []string

			BeforeEach(func() {
				expectedArgs = []string{
					filepath.Join(relativeDeploymentDir, "bosh.yml"),
					"--state", filepath.Join(relativeVarsDir, "bosh-state.json"),
					"--vars-store", filepath.Join(relativeVarsDir, "director-vars-store.yml"),
					"--vars-file", filepath.Join(relativeVarsDir, "director-vars-file.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "azure", "cpi.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "jumpbox-user.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "uaa.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "credhub.yml"),
				}
			})

			It("writes create-director.sh and delete-director.sh", func() {
				behavesLikePlan(expectedArgs, cmd, executor, dirInput, deploymentDir, "azure", stateDir)
			})
		})

		Context("vsphere", func() {
			var expectedArgs []string

			BeforeEach(func() {
				expectedArgs = []string{
					filepath.Join(relativeDeploymentDir, "bosh.yml"),
					"--state", filepath.Join(relativeVarsDir, "bosh-state.json"),
					"--vars-store", filepath.Join(relativeVarsDir, "director-vars-store.yml"),
					"--vars-file", filepath.Join(relativeVarsDir, "director-vars-file.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "vsphere", "cpi.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "jumpbox-user.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "uaa.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "credhub.yml"),
					"-o", filepath.Join(relativeDeploymentDir, "vsphere", "resource-pool.yml"),
				}
			})

			It("writes create-director.sh and delete-director.sh", func() {
				behavesLikePlan(expectedArgs, cmd, executor, dirInput, deploymentDir, "vsphere", stateDir)
			})
		})
	})

	Describe("WriteDeploymentVars", func() {
		var (
			executor bosh.Executor
			varsDir  string
			dirInput bosh.DirInput
		)

		BeforeEach(func() {
			var err error
			cmd := &fakes.BOSHCommand{}
			varsDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())
			stateDir, err := ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			executor = bosh.NewExecutor(cmd, ioutil.ReadFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)

			dirInput = bosh.DirInput{
				Deployment: "some-deployment",
				StateDir:   stateDir,
				VarsDir:    varsDir,
			}
		})

		It("writes the deployment vars yml file", func() {
			By("writing deployment vars to the state dir", func() {
				err := executor.WriteDeploymentVars(dirInput, "some-deployment-vars")
				Expect(err).NotTo(HaveOccurred())
				deploymentVars, err := ioutil.ReadFile(filepath.Join(varsDir, "some-deployment-vars-file.yml"))
				Expect(err).NotTo(HaveOccurred())

				Expect(string(deploymentVars)).To(Equal("some-deployment-vars"))
			})
		})
	})

	Describe("CreateEnv", func() {
		var (
			cmd      *fakes.BOSHCommand
			executor bosh.Executor

			createEnvPath string
			varsDir       string
			stateDir      string

			dirInput bosh.DirInput
			state    storage.State
		)

		BeforeEach(func() {
			var err error

			cmd = &fakes.BOSHCommand{}
			varsDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())
			stateDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			executor = bosh.NewExecutor(cmd, ioutil.ReadFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)

			dirInput = bosh.DirInput{
				Deployment: "some-deployment",
				StateDir:   stateDir,
				VarsDir:    varsDir,
			}

			state = storage.State{
				IAAS: "some-iaas",
			}

			createEnvPath = filepath.Join(stateDir, "create-some-deployment.sh")
			createEnvContents := fmt.Sprintf("#!/bin/bash\necho 'some-vars-store-contents' > %s/some-deployment-vars-store.yml\n", varsDir)

			ioutil.WriteFile(createEnvPath, []byte(createEnvContents), storage.ScriptMode)
		})

		AfterEach(func() {
			os.Remove(filepath.Join(varsDir, "some-deployment-vars-store.yml"))
			os.Remove(createEnvPath)
			os.Remove(filepath.Join(stateDir, "create-some-deployment-override.sh"))
			os.Unsetenv("BBL_STATE_DIR")
		})

		Context("when the user provides a create-env override", func() {
			BeforeEach(func() {
				overridePath := filepath.Join(stateDir, "create-some-deployment-override.sh")
				overrideContents := fmt.Sprintf("#!/bin/bash\necho 'override-vars-store-contents' > %s/some-deployment-vars-store.yml\n", varsDir)

				ioutil.WriteFile(overridePath, []byte(overrideContents), storage.ScriptMode)
			})

			It("runs the create-env-override.sh script", func() {
				vars, err := executor.CreateEnv(dirInput, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(cmd.RunCallCount()).To(Equal(0))
				Expect(vars).To(ContainSubstring("override-vars-store-contents"))
			})
		})

		It("runs the create-env script and returns the resulting vars-store contents", func() {
			vars, err := executor.CreateEnv(dirInput, state)
			Expect(err).NotTo(HaveOccurred())

			Expect(cmd.RunCallCount()).To(Equal(0))
			Expect(vars).To(ContainSubstring("some-vars-store-contents"))

			By("setting BBL_STATE_DIR environment variable", func() {
				bblStateDirEnv := os.Getenv("BBL_STATE_DIR")
				Expect(bblStateDirEnv).To(Equal(stateDir))
			})
		})

		Context("when the create-env script returns an error", func() {
			BeforeEach(func() {
				createEnvContents := "#!/bin/bash\nexit 1\n"
				ioutil.WriteFile(createEnvPath, []byte(createEnvContents), storage.ScriptMode)
			})

			It("returns an error", func() {
				vars, err := executor.CreateEnv(dirInput, state)
				Expect(err).To(MatchError("Run bosh create-env: exit status 1"))
				Expect(vars).To(Equal(""))
			})
		})
	})

	Describe("DeleteEnv", func() {
		var (
			cmd      *fakes.BOSHCommand
			executor bosh.Executor

			deleteEnvPath string
			varsDir       string
			stateDir      string

			dirInput bosh.DirInput
			state    storage.State
		)

		BeforeEach(func() {
			var err error
			cmd = &fakes.BOSHCommand{}
			varsDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())
			stateDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			executor = bosh.NewExecutor(cmd, ioutil.ReadFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)

			dirInput = bosh.DirInput{
				Deployment: "some-deployment",
				VarsDir:    varsDir,
				StateDir:   stateDir,
			}

			state = storage.State{
				IAAS: "some-iaas",
			}

			deleteEnvPath = filepath.Join(stateDir, "delete-some-deployment.sh")
			deleteEnvContents := "#!/bin/bash\necho delete-env > /dev/null\n"

			ioutil.WriteFile(deleteEnvPath, []byte(deleteEnvContents), storage.ScriptMode)
		})

		AfterEach(func() {
			os.Unsetenv("BBL_STATE_DIR")
			os.Remove(filepath.Join(stateDir, "delete-some-deployment.sh"))
		})

		Context("when the user provides a delete-env override", func() {
			BeforeEach(func() {
				overridePath := filepath.Join(stateDir, "delete-some-deployment-override.sh")
				overrideContents := fmt.Sprintf("#!/bin/bash\necho 'override' > %s/delete-env-output\n", varsDir)

				ioutil.WriteFile(overridePath, []byte(overrideContents), storage.ScriptMode)
			})

			AfterEach(func() {
				os.Remove(filepath.Join(varsDir, "delete-env-output"))
				os.Remove(filepath.Join(stateDir, "delete-some-deployment-override.sh"))
			})

			It("runs the delete-env-override.sh script", func() {
				err := executor.DeleteEnv(dirInput, state)
				Expect(err).NotTo(HaveOccurred())

				Expect(cmd.RunCallCount()).To(Equal(0))

				overrideOut, err := ioutil.ReadFile(filepath.Join(varsDir, "delete-env-output"))
				Expect(err).NotTo(HaveOccurred())
				Expect(overrideOut).To(ContainSubstring("override"))
			})
		})

		It("deletes a bosh environment with the delete-env script", func() {
			err := executor.DeleteEnv(dirInput, state)
			Expect(err).NotTo(HaveOccurred())

			Expect(cmd.RunCallCount()).To(Equal(0))

			By("setting BBL_STATE_DIR environment variable", func() {
				bblStateDirEnv := os.Getenv("BBL_STATE_DIR")
				Expect(bblStateDirEnv).To(Equal(stateDir))
			})
		})

		Context("when the create-env script returns an error", func() {
			BeforeEach(func() {
				deleteEnvContents := "#!/bin/bash\nexit 1\n"
				ioutil.WriteFile(deleteEnvPath, []byte(deleteEnvContents), storage.ScriptMode)
			})

			It("returns an error", func() {
				err := executor.DeleteEnv(dirInput, state)
				Expect(err).To(MatchError("Run bosh delete-env: exit status 1"))
			})
		})
	})

	Describe("Version", func() {
		var (
			cmd      *fakes.BOSHCommand
			executor bosh.Executor
		)
		BeforeEach(func() {
			cmd = &fakes.BOSHCommand{}
			cmd.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
				stdout.Write([]byte("some-text version 2.0.24 some-other-text"))
				return nil
			}

			executor = bosh.NewExecutor(cmd, ioutil.ReadFile, json.Unmarshal, json.Marshal, ioutil.WriteFile)
		})

		It("passes the correct args and dir to run command", func() {
			_, err := executor.Version()
			Expect(err).NotTo(HaveOccurred())

			_, _, args := cmd.RunArgsForCall(0)
			Expect(args).To(Equal([]string{"-v"}))
		})

		It("returns the correctly trimmed version", func() {
			version, err := executor.Version()
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal("2.0.24"))
		})

		Context("failure cases", func() {
			Context("when the run cmd fails", func() {
				BeforeEach(func() {
					cmd.RunReturns(errors.New("failed to run cmd"))
				})

				It("returns an error", func() {
					_, err := executor.Version()
					Expect(err).To(MatchError("failed to run cmd"))
				})
			})

			Context("when the version cannot be parsed", func() {
				var expectedError error

				BeforeEach(func() {
					expectedError = bosh.NewBOSHVersionError(errors.New("BOSH version could not be parsed"))
					cmd.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
						stdout.Write([]byte(""))
						return nil
					}
				})

				It("returns a bosh version error", func() {
					_, err := executor.Version()
					Expect(err).To(Equal(expectedError))
				})
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

func behavesLikePlan(expectedArgs []string, cmd *fakes.BOSHCommand, executor bosh.Executor, input bosh.DirInput, deploymentDir, iaas, stateDir string) {
	cmd.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
		stdout.Write([]byte("some-manifest"))
		return nil
	}

	err := executor.PlanDirector(input, deploymentDir, iaas)
	Expect(err).NotTo(HaveOccurred())
	Expect(cmd.RunCallCount()).To(Equal(0))

	By("writing the create-env args to a shell script", func() {
		expectedScript := formatScript("create-env", stateDir, expectedArgs)
		scriptPath := fmt.Sprintf("%s/create-director.sh", stateDir)
		shellScript, err := ioutil.ReadFile(scriptPath)
		Expect(err).NotTo(HaveOccurred())

		fileinfo, err := os.Stat(scriptPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(fileinfo.Mode().String()).To(Equal("-rwxr-x---"))
		Expect(err).NotTo(HaveOccurred())

		Expect(string(shellScript)).To(Equal(expectedScript))
	})

	By("writing the delete-env args to a shell script", func() {
		expectedScript := formatScript("delete-env", stateDir, expectedArgs)
		scriptPath := fmt.Sprintf("%s/delete-director.sh", stateDir)
		shellScript, err := ioutil.ReadFile(scriptPath)
		Expect(err).NotTo(HaveOccurred())

		fileinfo, err := os.Stat(scriptPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(fileinfo.Mode().String()).To(Equal("-rwxr-x---"))
		Expect(err).NotTo(HaveOccurred())
		Expect(err).NotTo(HaveOccurred())

		Expect(string(shellScript)).To(Equal(expectedScript))
	})
}
