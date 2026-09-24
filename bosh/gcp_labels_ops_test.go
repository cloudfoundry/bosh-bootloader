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
