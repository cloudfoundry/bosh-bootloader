package config_test

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/config"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GCP labels", func() {
	var (
		merger      config.Merger
		state       storage.State
		globalFlags config.GlobalFlags
		mergedState storage.State
		mergeErr    error
	)

	BeforeEach(func() {
		merger = config.NewMerger(&fakes.FileIO{})
		state = storage.State{IAAS: "gcp"}
		globalFlags = config.GlobalFlags{IAAS: "gcp"}
	})

	JustBeforeEach(func() {
		mergedState, mergeErr = merger.MergeGlobalFlagsToState(globalFlags, state)
	})

	Context("when no labels are provided", func() {
		It("leaves the labels unset", func() {
			Expect(mergeErr).NotTo(HaveOccurred())
			Expect(mergedState.GCP.Labels).To(BeNil())
		})
	})

	Context("when labels are provided", func() {
		BeforeEach(func() {
			globalFlags.GCPLabels = []string{"pipeline=bosh-deployment", "Team=FIWG"}
		})

		It("stores them lower-cased", func() {
			Expect(mergeErr).NotTo(HaveOccurred())
			Expect(mergedState.GCP.Labels).To(Equal(map[string]string{
				"pipeline": "bosh-deployment",
				"team":     "fiwg",
			}))
		})
	})

	Context("when a label key contains an underscore", func() {
		BeforeEach(func() {
			globalFlags.GCPLabels = []string{"cost_center=platform"}
		})

		It("accepts it", func() {
			Expect(mergeErr).NotTo(HaveOccurred())
			Expect(mergedState.GCP.Labels).To(Equal(map[string]string{
				"cost_center": "platform",
			}))
		})
	})

	Context("when a label is missing a value separator", func() {
		BeforeEach(func() {
			globalFlags.GCPLabels = []string{"pipeline"}
		})

		It("returns an error", func() {
			Expect(mergeErr).To(MatchError(ContainSubstring(`invalid GCP label "pipeline": expected key=value`)))
		})
	})

	Context("when a label key is invalid", func() {
		BeforeEach(func() {
			globalFlags.GCPLabels = []string{"bad key=value"}
		})

		It("returns an error", func() {
			Expect(mergeErr).To(MatchError(ContainSubstring(`invalid GCP label key "bad key"`)))
		})
	})

	Context("when a label value is invalid", func() {
		BeforeEach(func() {
			globalFlags.GCPLabels = []string{"pipeline=bad value"}
		})

		It("returns an error", func() {
			Expect(mergeErr).To(MatchError(ContainSubstring(`invalid GCP label value "bad value"`)))
		})
	})

	Context("when a label value is empty", func() {
		BeforeEach(func() {
			globalFlags.GCPLabels = []string{"pipeline="}
		})

		It("returns an error", func() {
			Expect(mergeErr).To(MatchError(ContainSubstring(`invalid GCP label value ""`)))
		})
	})

	Context("when an entry is blank", func() {
		BeforeEach(func() {
			globalFlags.GCPLabels = []string{""}
		})

		It("ignores it and leaves the labels unset", func() {
			Expect(mergeErr).NotTo(HaveOccurred())
			Expect(mergedState.GCP.Labels).To(BeNil())
		})
	})

	Context("when more than 64 labels are provided", func() {
		BeforeEach(func() {
			globalFlags.GCPLabels = make([]string, 65)
			for i := range globalFlags.GCPLabels {
				globalFlags.GCPLabels[i] = fmt.Sprintf("key%d=value", i)
			}
		})

		It("returns an error", func() {
			Expect(mergeErr).To(MatchError(ContainSubstring("too many GCP labels")))
		})
	})
})
