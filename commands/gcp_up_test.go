package commands_test

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("gcp up", func() {
	var (
		stateStore        *fakes.StateStore
		keyPairUpdater    *fakes.GCPKeyPairUpdater
		gcpUp             commands.GCPUp
		gcpClientProvider *fakes.GCPClientProvider

		serviceAccountKeyPath string
		serviceAccountKey     string
		terraformApplier      *fakes.TerraformApplier
	)

	BeforeEach(func() {
		stateStore = &fakes.StateStore{}
		keyPairUpdater = &fakes.GCPKeyPairUpdater{}
		gcpClientProvider = &fakes.GCPClientProvider{}
		terraformApplier = &fakes.TerraformApplier{}

		gcpUp = commands.NewGCPUp(stateStore, keyPairUpdater, gcpClientProvider, terraformApplier)

		tempFile, err := ioutil.TempFile("", "gcpServiceAccountKey")
		Expect(err).NotTo(HaveOccurred())

		serviceAccountKeyPath = tempFile.Name()
		serviceAccountKey = `{"real": "json"}`
		err = ioutil.WriteFile(serviceAccountKeyPath, []byte(serviceAccountKey), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Execute", func() {
		It("saves gcp details to the state", func() {
			keyPairUpdater.UpdateCall.Returns.KeyPair = storage.KeyPair{
				PrivateKey: "some-private-key",
				PublicKey:  "some-public-key",
			}

			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKeyPath: serviceAccountKeyPath,
				ProjectID:             "some-project-id",
				Zone:                  "some-zone",
				Region:                "some-region",
			}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(stateStore.SetCall.Receives.State).To(Equal(storage.State{
				IAAS: "gcp",
				GCP: storage.GCP{
					ServiceAccountKey: serviceAccountKey,
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "some-region",
				},
				KeyPair: storage.KeyPair{
					PrivateKey: "some-private-key",
					PublicKey:  "some-public-key",
				},
			}))
			Expect(gcpClientProvider.SetConfigCall.CallCount).To(Equal(1))
			Expect(gcpClientProvider.SetConfigCall.Receives.ServiceAccountKey).To(Equal(`{"real": "json"}`))
		})

		It("uploads the ssh keys", func() {
			err := gcpUp.Execute(commands.GCPUpConfig{
				ServiceAccountKeyPath: serviceAccountKeyPath,
				ProjectID:             "some-project-id",
				Zone:                  "some-zone",
				Region:                "some-region",
			}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(keyPairUpdater.UpdateCall.CallCount).To(Equal(1))
			Expect(keyPairUpdater.UpdateCall.Receives.ProjectID).To(Equal("some-project-id"))
		})

		It("creates gcp resources via terraform", func() {
			terraformApplier.ApplyCall.Returns.TFState = "my-tf-state"
			gcpUpConfig := commands.GCPUpConfig{
				ServiceAccountKeyPath: serviceAccountKeyPath,
				ProjectID:             "some-project-id",
				Zone:                  "some-zone",
				Region:                "some-region",
			}

			err := gcpUp.Execute(gcpUpConfig, storage.State{
				EnvID: "some-env-id",
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(terraformApplier.ApplyCall.CallCount).To(Equal(1))
			Expect(terraformApplier.ApplyCall.Receives.Credentials).To(Equal(serviceAccountKeyPath))
			Expect(terraformApplier.ApplyCall.Receives.EnvID).To(Equal("some-env-id"))
			Expect(terraformApplier.ApplyCall.Receives.ProjectID).To(Equal("some-project-id"))
			Expect(terraformApplier.ApplyCall.Receives.Zone).To(Equal("some-zone"))
			Expect(terraformApplier.ApplyCall.Receives.Region).To(Equal("some-region"))
			Expect(terraformApplier.ApplyCall.Receives.Template).To(Equal(`variable "project_id" {
	type = "string"
}

variable "region" {
	type = "string"
}

variable "zone" {
	type = "string"
}

variable "env_id" {
	type = "string"
}

variable "credentials" {
	type = "string"
}

provider "google" {
	credentials = "${file("${var.credentials}")}"
	project = "${var.project_id}"
	region = "${var.region}"
}

resource "google_compute_network" "bbl-network" {
  name		 = "${var.env_id}-network"
}

resource "google_compute_subnetwork" "bbl-subnet" {
  name			= "${var.env_id}-subnet"
  ip_cidr_range = "10.0.0.0/16"
  network		= "${google_compute_network.bbl-network.self_link}"
}

resource "google_compute_address" "bosh-external-ip" {
  name = "${var.env_id}-bosh-external-ip"
}

resource "google_compute_firewall" "bosh-open" {
  name    = "${var.env_id}-bosh-open"
  network = "${google_compute_network.bbl-network.name}"

  source_ranges = ["0.0.0.0/0"]

  allow {
    protocol = "icmp"
  }

  allow {
    ports = ["22", "6868", "25555"]
    protocol = "tcp"
  }

  target_tags = ["${var.env_id}-bosh-open"]
}

resource "google_compute_firewall" "internal" {
  name    = "${var.env_id}-internal"
  network = "${google_compute_network.bbl-network.name}"

  allow {
    protocol = "icmp"
  }

  allow {
    protocol = "tcp"
  }

  allow {
    protocol = "udp"
  }

  source_tags = ["${var.env_id}-bosh-open","${var.env_id}-internal"]
}`))
			Expect(stateStore.SetCall.Receives.State.TFState).To(Equal(`my-tf-state`))
		})

		Context("when state contains gcp details", func() {
			var (
				updatedServiceAccountKey     string
				updatedServiceAccountKeyPath string
			)

			BeforeEach(func() {
				tempFile, err := ioutil.TempFile("", "updatedGcpServiceAccountKey")
				Expect(err).NotTo(HaveOccurred())

				updatedServiceAccountKeyPath = tempFile.Name()
				updatedServiceAccountKey = `{"another-real": "json-file"}`
				err = ioutil.WriteFile(updatedServiceAccountKeyPath, []byte(updatedServiceAccountKey), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			It("overwrites them with the up config details", func() {
				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKeyPath: updatedServiceAccountKeyPath,
					ProjectID:             "new-project-id",
					Zone:                  "new-zone",
					Region:                "new-region",
				}, storage.State{
					IAAS: "gcp",
					GCP: storage.GCP{
						ServiceAccountKey: serviceAccountKey,
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "some-region",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(stateStore.SetCall.Receives.State).To(Equal(storage.State{
					IAAS: "gcp",
					GCP: storage.GCP{
						ServiceAccountKey: updatedServiceAccountKey,
						ProjectID:         "new-project-id",
						Zone:              "new-zone",
						Region:            "new-region",
					},
				}))
			})

			It("does not create a new ssh key", func() {
				err := gcpUp.Execute(commands.GCPUpConfig{}, storage.State{
					IAAS: "gcp",
					GCP: storage.GCP{
						ServiceAccountKey: serviceAccountKey,
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "some-region",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(keyPairUpdater.UpdateCall.CallCount).To(Equal(0))
			})

			It("calls terraform applier with previous tf state", func() {
				err := gcpUp.Execute(commands.GCPUpConfig{}, storage.State{
					IAAS: "gcp",
					GCP: storage.GCP{
						ServiceAccountKey: serviceAccountKey,
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "some-region",
					},
					TFState: "some-tf-state",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(terraformApplier.ApplyCall.CallCount).To(Equal(1))
				Expect(terraformApplier.ApplyCall.Receives.TFState).To(Equal("some-tf-state"))
			})

			It("does not require details from up config", func() {
				err := gcpUp.Execute(commands.GCPUpConfig{}, storage.State{
					IAAS: "gcp",
					GCP: storage.GCP{
						ServiceAccountKey: serviceAccountKey,
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "some-region",
					},
				})
				Expect(err).NotTo(HaveOccurred())
			})

			DescribeTable("up config contains subset of the details", func(upConfig commands.GCPUpConfig, expectedErr string) {
				err := gcpUp.Execute(upConfig, storage.State{
					IAAS: "gcp",
					GCP: storage.GCP{
						ServiceAccountKey: serviceAccountKey,
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "some-region",
					},
				})
				Expect(err).To(MatchError(expectedErr))
			},
				Entry("returns an error when the service account key is not provided", commands.GCPUpConfig{
					ProjectID: "new-project-id",
					Zone:      "new-zone",
					Region:    "new-region",
				}, "GCP service account key must be provided"),
				Entry("returns an error when the project ID is not provided", commands.GCPUpConfig{
					ServiceAccountKeyPath: "new-service-account-key",
					Zone:   "new-zone",
					Region: "new-region",
				}, "GCP project ID must be provided"),
				Entry("returns an error when the zone is not provided", commands.GCPUpConfig{
					ServiceAccountKeyPath: "new-service-account-key",
					ProjectID:             "new-project-id",
					Region:                "new-region",
				}, "GCP zone must be provided"),
				Entry("returns an error when the region is not provided", commands.GCPUpConfig{
					ServiceAccountKeyPath: "new-service-account-key",
					ProjectID:             "new-project-id",
					Zone:                  "new-zone",
				}, "GCP region must be provided"),
			)
		})

		Context("failure cases", func() {
			It("returns an error when state store fails", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{Error: errors.New("set call failed")}}
				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKeyPath: serviceAccountKeyPath,
					ProjectID:             "p",
					Zone:                  "z",
					Region:                "r",
				}, storage.State{})
				Expect(err).To(MatchError("set call failed"))
			})

			DescribeTable("up config validation", func(upConfig commands.GCPUpConfig, expectedErr string) {
				err := gcpUp.Execute(upConfig, storage.State{})
				Expect(err).To(MatchError(expectedErr))
			},
				Entry("returns an error when service account key is missing", commands.GCPUpConfig{
					ProjectID: "p",
					Zone:      "z",
					Region:    "r",
				}, "GCP service account key must be provided"),
				Entry("returns an error when project ID is missing", commands.GCPUpConfig{
					ServiceAccountKeyPath: "sak",
					Zone:   "z",
					Region: "r",
				}, "GCP project ID must be provided"),
				Entry("returns an error when zone is missing", commands.GCPUpConfig{
					ServiceAccountKeyPath: "sak",
					ProjectID:             "p",
					Region:                "r",
				}, "GCP zone must be provided"),
				Entry("returns an error when region is missing", commands.GCPUpConfig{
					ServiceAccountKeyPath: "sak",
					ProjectID:             "p",
					Zone:                  "z",
				}, "GCP region must be provided"),
			)

			It("returns an error when the service account key file does not exist", func() {
				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKeyPath: "/some/non/existent/file",
					ProjectID:             "p",
					Zone:                  "z",
					Region:                "r",
				}, storage.State{})
				Expect(err).To(MatchError("error reading service account key: open /some/non/existent/file: no such file or directory"))
			})

			It("returns an error when the service account key file does not contain valid json", func() {
				tempFile, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())

				invalidServiceAccountKeyPath := tempFile.Name()
				err = ioutil.WriteFile(invalidServiceAccountKeyPath, []byte(`%%%not-valid-json%%%`), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKeyPath: invalidServiceAccountKeyPath,
					ProjectID:             "p",
					Zone:                  "z",
					Region:                "r",
				}, storage.State{})
				Expect(err).To(MatchError("error parsing service account key: invalid character '%' looking for beginning of value"))
			})

			It("returns an error when the keypair could not be updated", func() {
				keyPairUpdater.UpdateCall.Returns.Error = errors.New("keypair update failed")

				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKeyPath: serviceAccountKeyPath,
					ProjectID:             "some-project-id",
					Zone:                  "some-zone",
					Region:                "some-region",
				}, storage.State{})
				Expect(err).To(MatchError("keypair update failed"))
			})

			It("returns an error when setting config fails", func() {
				gcpClientProvider.SetConfigCall.Returns.Error = errors.New("setting config failed")

				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKeyPath: serviceAccountKeyPath,
					ProjectID:             "some-project-id",
					Zone:                  "some-zone",
					Region:                "some-region",
				}, storage.State{})
				Expect(err).To(MatchError("setting config failed"))
			})

			It("returns an error when terraform applier fails", func() {
				terraformApplier.ApplyCall.Returns.Error = errors.New("terraform applier failed")

				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKeyPath: serviceAccountKeyPath,
					ProjectID:             "some-project-id",
					Zone:                  "some-zone",
					Region:                "some-region",
				}, storage.State{})
				Expect(err).To(MatchError("terraform applier failed"))
			})

			It("returns an error when the state fails to be set", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {errors.New("state failed to be set")}}

				err := gcpUp.Execute(commands.GCPUpConfig{
					ServiceAccountKeyPath: serviceAccountKeyPath,
					ProjectID:             "some-project-id",
					Zone:                  "some-zone",
					Region:                "some-region",
				}, storage.State{})
				Expect(err).To(MatchError("state failed to be set"))
			})
		})
	})
})
