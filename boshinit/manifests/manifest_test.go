package manifests_test

import (
	"github.com/cloudfoundry/bosh-bootloader/boshinit/manifests"
	"github.com/cloudfoundry/bosh-bootloader/ssl"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manifest", func() {
	Describe("DirectorSSLKeyPair", func() {
		It("returns the director ssl keypair from a built manifest", func() {
			manifest := manifests.Manifest{
				Jobs: []manifests.Job{{
					Properties: manifests.JobProperties{
						Director: manifests.DirectorJobProperties{
							SSL: manifests.SSLProperties{
								Cert: "some-cert",
								Key:  "some-key",
							},
						},
					},
				}},
			}

			keyPair := manifest.DirectorSSLKeyPair()
			Expect(keyPair).To(Equal(ssl.KeyPair{
				Certificate: []byte("some-cert"),
				PrivateKey:  []byte("some-key"),
			}))
		})

		It("returns an empty keypair if there are no jobs", func() {
			manifest := manifests.Manifest{}

			keyPair := manifest.DirectorSSLKeyPair()
			Expect(keyPair).To(Equal(ssl.KeyPair{}))
		})
	})
})
