package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Firewall", func() {
	var (
		client *fakes.FirewallsClient
		name   string

		firewall compute.Firewall
	)

	BeforeEach(func() {
		client = &fakes.FirewallsClient{}
		name = "banana"

		firewall = compute.NewFirewall(client, name)
	})

	Describe("Delete", func() {
		It("deletes the firewall", func() {
			err := firewall.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteFirewallCall.CallCount).To(Equal(1))
			Expect(client.DeleteFirewallCall.Receives.Firewall).To(Equal(name))
		})

		Context("when the client fails to delete", func() {
			BeforeEach(func() {
				client.DeleteFirewallCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := firewall.Delete()
				Expect(err).To(MatchError("ERROR deleting firewall banana: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(firewall.Name()).To(Equal(name))
		})
	})
})
