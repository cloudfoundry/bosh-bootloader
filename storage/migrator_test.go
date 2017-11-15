package storage_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Migrator", func() {
	var (
		migrator storage.Migrator
		store    *fakes.StateStore
		varsDir  string
	)

	BeforeEach(func() {
		store = &fakes.StateStore{}
		migrator = storage.NewMigrator(store)

		var err error
		varsDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		store.GetVarsDirCall.Returns.Directory = varsDir
	})

	Describe("Migrate", func() {
		var incomingState storage.State

		Context("when the state is empty", func() {
			It("returns the state without changing it", func() {
				outgoingState, err := migrator.Migrate(storage.State{})
				Expect(err).NotTo(HaveOccurred())
				Expect(outgoingState).To(Equal(storage.State{}))
				Expect(store.SetCall.CallCount).To(Equal(0))
			})
		})

		Context("when the state is already migrated", func() {
			BeforeEach(func() {
				incomingState = storage.State{EnvID: "some-env-id"}
			})

			Context("when the vars dir cannot be retrieved", func() {
				BeforeEach(func() {
					store.GetVarsDirCall.Returns.Error = errors.New("potato")
				})

				It("returns an error", func() {
					_, err := migrator.Migrate(incomingState)
					Expect(err).To(MatchError("migrating state: potato"))
				})
			})

			Context("when the state cannot be saved", func() {
				BeforeEach(func() {
					store.SetCall.Returns = []fakes.SetCallReturn{
						fakes.SetCallReturn{
							Error: errors.New("tomato"),
						},
					}
				})

				It("returns an error", func() {
					_, err := migrator.Migrate(incomingState)
					Expect(err).To(MatchError("saving migrated state: tomato"))
				})
			})

		})

		Context("when the state has a populated TFState", func() {
			BeforeEach(func() {
				incomingState = storage.State{
					EnvID:   "some-env-id",
					TFState: "some-tf-state",
				}
			})

			It("writes the TFState to the tfstate file", func() {
				outgoingState, err := migrator.Migrate(incomingState)
				Expect(err).NotTo(HaveOccurred())
				Expect(outgoingState.TFState).To(BeEmpty())

				contents, err := ioutil.ReadFile(filepath.Join(varsDir, "terraform.tfstate"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal("some-tf-state"))

				By("saving the state after removing the old values", func() {
					Expect(store.SetCall.CallCount).To(Equal(1))
					Expect(store.SetCall.Receives[0].State).To(Equal(storage.State{EnvID: "some-env-id"}))
				})

				By("not writing the bosh state", func() {
					_, err := os.Stat(filepath.Join(varsDir, "bosh-state.json"))
					Expect(err).To(HaveOccurred())
				})
			})

			Context("failure cases", func() {
				Context("when the tfstate file cannot be written", func() {
					BeforeEach(func() {
						err := os.MkdirAll(filepath.Join(varsDir, "terraform.tfstate"), os.ModePerm)
						Expect(err).NotTo(HaveOccurred())
					})

					It("returns an error", func() {
						_, err := migrator.Migrate(incomingState)
						Expect(err).To(MatchError(ContainSubstring("migrating terraform state: ")))
					})
				})
			})
		})

		Context("when the state has a populated BOSH state", func() {
			BeforeEach(func() {
				incomingState = storage.State{
					EnvID: "some-env-id",
					BOSH: storage.BOSH{
						DirectorAddress: "10.0.0.6",
						State: map[string]interface{}{
							"some-bosh-key": "some-bosh-value",
						},
					},
				}
			})
			It("copies the BOSH state to the bosh-state.json vars file", func() {
				_, err := migrator.Migrate(incomingState)
				Expect(err).NotTo(HaveOccurred())
				boshState, err := ioutil.ReadFile(filepath.Join(varsDir, "bosh-state.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(boshState).To(MatchJSON(`{"some-bosh-key": "some-bosh-value"}`))

				By("saving the state after removing the old values", func() {
					Expect(store.SetCall.CallCount).To(Equal(1))
					Expect(store.SetCall.Receives[0].State).To(Equal(storage.State{
						EnvID: "some-env-id",
						BOSH: storage.BOSH{
							DirectorAddress: "10.0.0.6",
						},
					}))
				})
			})
			Context("failure cases", func() {
				Context("when the bosh state file cannot be written", func() {
					BeforeEach(func() {
						err := os.MkdirAll(filepath.Join(varsDir, "bosh-state.json"), os.ModePerm)
						Expect(err).NotTo(HaveOccurred())
					})

					It("returns an error", func() {
						_, err := migrator.Migrate(incomingState)
						Expect(err).To(MatchError(ContainSubstring("migrating bosh state: ")))
					})
				})
				Context("when the bosh state file cannot be written", func() {
					BeforeEach(func() {
						incomingState.BOSH.State["invalid-key"] = func() string { return "invalid" }
					})

					It("returns an error", func() {
						_, err := migrator.Migrate(incomingState)
						Expect(err).To(MatchError(ContainSubstring("marshalling bosh state: ")))
					})
				})
			})

			Context("when the state has populated BOSH variables", func() {
				BeforeEach(func() {
					incomingState = storage.State{
						EnvID: "some-env-id",
						BOSH: storage.BOSH{
							DirectorAddress: "10.0.0.6",
							Variables:       "some-bosh-vars",
						},
					}
				})
				It("copies the BOSH state to the director-variables.yml file", func() {
					_, err := migrator.Migrate(incomingState)
					Expect(err).NotTo(HaveOccurred())
					boshVars, err := ioutil.ReadFile(filepath.Join(varsDir, "director-variables.yml"))
					Expect(err).NotTo(HaveOccurred())
					Expect(string(boshVars)).To(Equal("some-bosh-vars"))

					By("saving the state after removing the old values", func() {
						Expect(store.SetCall.CallCount).To(Equal(1))
						Expect(store.SetCall.Receives[0].State).To(Equal(storage.State{
							EnvID: "some-env-id",
							BOSH: storage.BOSH{
								DirectorAddress: "10.0.0.6",
							},
						}))
					})
				})
				Context("failure cases", func() {
					Context("when the director variables file cannot be written", func() {
						BeforeEach(func() {
							err := os.MkdirAll(filepath.Join(varsDir, "director-variables.yml"), os.ModePerm)
							Expect(err).NotTo(HaveOccurred())
						})

						It("returns an error", func() {
							_, err := migrator.Migrate(incomingState)
							Expect(err).To(MatchError(ContainSubstring("migrating bosh variables: ")))
						})
					})
				})
			})

			Context("when the state has populated jumpbox variables", func() {
				BeforeEach(func() {
					incomingState = storage.State{
						EnvID: "some-env-id",
						Jumpbox: storage.Jumpbox{
							URL:       "10.0.0.5:25555",
							Variables: "some-jumpbox-vars",
						},
					}
				})
				It("copies the jumpbox state to the jumpbox-variables.yml file", func() {
					_, err := migrator.Migrate(incomingState)
					Expect(err).NotTo(HaveOccurred())
					jumpboxVars, err := ioutil.ReadFile(filepath.Join(varsDir, "jumpbox-variables.yml"))
					Expect(err).NotTo(HaveOccurred())
					Expect(string(jumpboxVars)).To(Equal("some-jumpbox-vars"))

					By("saving the state after removing the old values", func() {
						Expect(store.SetCall.CallCount).To(Equal(1))
						Expect(store.SetCall.Receives[0].State).To(Equal(storage.State{
							EnvID: "some-env-id",
							Jumpbox: storage.Jumpbox{
								URL: "10.0.0.5:25555",
							},
						}))
					})
				})
				Context("failure cases", func() {
					Context("when the director variables file cannot be written", func() {
						BeforeEach(func() {
							err := os.MkdirAll(filepath.Join(varsDir, "jumpbox-variables.yml"), os.ModePerm)
							Expect(err).NotTo(HaveOccurred())
						})

						It("returns an error", func() {
							_, err := migrator.Migrate(incomingState)
							Expect(err).To(MatchError(ContainSubstring("migrating jumpbox variables: ")))
						})
					})
				})
			})
		})
	})
})
