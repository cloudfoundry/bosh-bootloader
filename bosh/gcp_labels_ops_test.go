package bosh_test

import (
	"github.com/cloudfoundry/bosh-bootloader/bosh"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GCPLabelsOps", func() {
	It("renders an ops file that sets labels on the director and jumpbox cloud properties", func() {
		ops, err := bosh.GCPLabelsOps(map[string]string{
			"pipeline": "bosh-deployment",
			"owner":    "fiwg",
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(string(ops)).To(Equal(`- type: replace
  path: /resource_pools/name=vms/cloud_properties/labels?
  value:
    owner: fiwg
    pipeline: bosh-deployment
`))
	})
})

var _ = Describe("GCPLabelsRuntimeConfigOps", func() {
	It("renders one replace operation per label so unrelated tags are preserved", func() {
		ops, err := bosh.GCPLabelsRuntimeConfigOps(map[string]string{
			"pipeline": "bosh-deployment",
			"owner":    "fiwg",
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(string(ops)).To(Equal(`- type: replace
  path: /tags?/owner?
  value: fiwg
- type: replace
  path: /tags?/pipeline?
  value: bosh-deployment
`))
	})
})
