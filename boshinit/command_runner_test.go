package boshinit_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/boshinit"
	"github.com/cloudfoundry/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CommandRunner", func() {
	Describe("Execute", func() {
		var (
			tempDir    string
			executable *fakes.Executable
			runner     boshinit.CommandRunner
		)

		BeforeEach(func() {
			var err error
			tempDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			executable = &fakes.Executable{}

			executable.RunCall.Stub = func() error {
				return ioutil.WriteFile(filepath.Join(tempDir, "bosh-state.json"), []byte(`{"key": "value"}`), os.ModePerm)
			}

			runner = boshinit.NewCommandRunner(tempDir, executable)
		})

		It("writes out the bosh.yml file to a temporary directory", func() {
			_, err := runner.Execute([]byte("some-manifest-yaml"), "some-private-key", boshinit.State{})
			Expect(err).NotTo(HaveOccurred())

			manifest, err := ioutil.ReadFile(filepath.Join(tempDir, "bosh.yml"))
			Expect(err).NotTo(HaveOccurred())
			Expect(manifest).To(ContainSubstring("some-manifest-yaml"))

			fileInfo, err := os.Stat(filepath.Join(tempDir, "bosh.yml"))
			Expect(err).NotTo(HaveOccurred())
			Expect(fileInfo.Mode()).To(Equal(os.FileMode(0644)))
		})

		It("writes out the private key", func() {
			_, err := runner.Execute([]byte("some-manifest-yaml"), "some-private-key", boshinit.State{})
			Expect(err).NotTo(HaveOccurred())

			manifest, err := ioutil.ReadFile(filepath.Join(tempDir, "bosh.pem"))
			Expect(err).NotTo(HaveOccurred())
			Expect(manifest).To(ContainSubstring("some-private-key"))

			fileInfo, err := os.Stat(filepath.Join(tempDir, "bosh.pem"))
			Expect(err).NotTo(HaveOccurred())
			Expect(fileInfo.Mode()).To(Equal(os.FileMode(0644)))
		})

		It("runs the executable", func() {
			_, err := runner.Execute([]byte("some-manifest-yaml"), "some-private-key", boshinit.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(executable.RunCall.CallCount).To(Equal(1))
		})

		Context("when the bosh-state.json file exists", func() {
			It("returns a bosh state object", func() {
				state, err := runner.Execute([]byte("some-manifest-yaml"), "some-private-key", boshinit.State{})
				Expect(err).NotTo(HaveOccurred())
				Expect(state).To(Equal(boshinit.State{
					"key": "value",
				}))
			})
		})

		Context("when the bosh-state.json file does not exist", func() {
			It("returns an empty bosh state object", func() {
				executable.RunCall.Stub = func() error {
					return os.Remove(filepath.Join(tempDir, "bosh-state.json"))
				}

				state, err := runner.Execute([]byte("some-manifest-yaml"), "some-private-key", boshinit.State{})
				Expect(err).NotTo(HaveOccurred())
				Expect(state).To(Equal(boshinit.State{}))
			})
		})

		It("receives a bosh state object and writes the bosh-state.json file", func() {
			executable.RunCall.Stub = nil
			_, err := runner.Execute([]byte("some-manifest-yaml"), "some-private-key", boshinit.State{
				"original_key": "original_value",
			})
			Expect(err).NotTo(HaveOccurred())

			file, err := ioutil.ReadFile(filepath.Join(tempDir, "bosh-state.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(file).To(MatchJSON(`{
				"original_key": "original_value"
			}`))

			fileInfo, err := os.Stat(filepath.Join(tempDir, "bosh-state.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(fileInfo.Mode()).To(Equal(os.FileMode(0644)))
		})

		Context("failure cases", func() {
			Context("when the bosh-init state cannot be marshaled", func() {
				It("returns an error", func() {
					_, err := runner.Execute([]byte("some-manifest-yaml"), "some-private-key", boshinit.State{
						"key": func() {},
					})
					Expect(err).To(MatchError(ContainSubstring("unsupported type: func()")))
				})
			})

			Context("when the bosh-state.json file cannot be written", func() {
				It("returns an error", func() {
					err := ioutil.WriteFile(filepath.Join(tempDir, "bosh-state.json"), []byte(""), os.ModePerm)
					Expect(err).NotTo(HaveOccurred())

					err = os.Chmod(filepath.Join(tempDir, "bosh-state.json"), os.FileMode(0000))
					Expect(err).NotTo(HaveOccurred())

					_, err = runner.Execute([]byte("some-manifest-yaml"), "some-private-key", boshinit.State{})
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			Context("when the bosh.yml file write fails", func() {
				It("returns an error", func() {
					err := ioutil.WriteFile(filepath.Join(tempDir, "bosh.yml"), []byte(""), os.ModePerm)
					Expect(err).NotTo(HaveOccurred())

					err = os.Chmod(filepath.Join(tempDir, "bosh.yml"), os.FileMode(0000))
					Expect(err).NotTo(HaveOccurred())

					_, err = runner.Execute([]byte("some-manifest-yaml"), "some-private-key", boshinit.State{})
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			Context("when the bosh.pem file write fails", func() {
				It("returns an error", func() {
					err := ioutil.WriteFile(filepath.Join(tempDir, "bosh.pem"), []byte(""), os.ModePerm)
					Expect(err).NotTo(HaveOccurred())

					err = os.Chmod(filepath.Join(tempDir, "bosh.pem"), os.FileMode(0000))
					Expect(err).NotTo(HaveOccurred())

					_, err = runner.Execute([]byte("some-manifest-yaml"), "some-private-key", boshinit.State{})
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			Context("when the command fails to run", func() {
				It("returns an error", func() {
					executable.RunCall.Stub = nil
					executable.RunCall.Returns.Error = errors.New("failed to run")

					_, err := runner.Execute([]byte("some-manifest-yaml"), "some-private-key", boshinit.State{})
					Expect(err).To(MatchError("failed to run"))
				})
			})

			Context("when bosh-state.json cannot be read", func() {
				It("returns an error", func() {
					executable.RunCall.Stub = func() error {
						err := ioutil.WriteFile(filepath.Join(tempDir, "bosh-state.json"), []byte(`{"key": "value"}`), os.ModePerm)
						if err != nil {
							return err
						}

						return os.Chmod(filepath.Join(tempDir, "bosh-state.json"), 0000)
					}

					_, err := runner.Execute([]byte("some-manifest-yaml"), "some-private-key", boshinit.State{})
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			Context("when bosh-state.json cannot be unmarshalled", func() {
				It("returns an error", func() {
					executable.RunCall.Stub = func() error {
						return ioutil.WriteFile(filepath.Join(tempDir, "bosh-state.json"), []byte("%%%%%"), os.ModePerm)
					}

					_, err := runner.Execute([]byte("some-manifest-yaml"), "some-private-key", boshinit.State{})
					Expect(err).To(MatchError(ContainSubstring("invalid character")))
				})
			})
		})
	})
})
