package cartographer_test

import (
	"io/ioutil"

	"github.com/genevieve/cartographer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

var _ = Describe("Cartographer", func() {
	var (
		tfstate string
		carto   cartographer.Cartographer
	)

	BeforeEach(func() {
		tfstate = "fixtures/terraform.tfstate"
		carto = cartographer.NewCartographer()
	})

	Describe("Ymlize", func() {
		var expectedYml string

		BeforeEach(func() {
			cont, err := ioutil.ReadFile("fixtures/all-vars.yml")
			Expect(err).NotTo(HaveOccurred())
			expectedYml = string(cont)
		})

		It("converts terraform outputs to yml", func() {
			yml, err := carto.Ymlize(tfstate)
			Expect(err).NotTo(HaveOccurred())
			Expect(yml).To(HelpfullyMatchYAML(expectedYml))
		})

		PContext("when there is a prefix", func() {})
		PContext("when no dest is provided", func() {})

		Context("when the source file path is invalid", func() {
			It("returns an error", func() {
				_, err := carto.Ymlize("banana")
				Expect(err).To(MatchError("open banana: no such file or directory"))
			})
		})

		Context("when the source file content is invalid", func() {
			It("returns an error", func() {
				_, err := carto.Ymlize("fixtures/invalid.tfstate")
				Expect(err).To(MatchError("invalid character '%' looking for beginning of object key string"))
			})
		})
	})

	Describe("YmlizeWithPrefix", func() {
		Context("for director prefix", func() {
			var expectedYml string

			BeforeEach(func() {
				cont, err := ioutil.ReadFile("fixtures/director-vars.yml")
				Expect(err).NotTo(HaveOccurred())
				expectedYml = string(cont)
			})

			It("converts terraform outputs to yml", func() {
				yml, err := carto.YmlizeWithPrefix(tfstate, "director")
				Expect(err).NotTo(HaveOccurred())
				Expect(yml).To(HelpfullyMatchYAML(expectedYml))
			})
		})

		Context("for jumpbox prefix", func() {
			var expectedYml string

			BeforeEach(func() {
				cont, err := ioutil.ReadFile("fixtures/jumpbox-vars.yml")
				Expect(err).NotTo(HaveOccurred())
				expectedYml = string(cont)
			})

			It("converts terraform outputs to yml", func() {
				yml, err := carto.YmlizeWithPrefix(tfstate, "jumpbox")
				Expect(err).NotTo(HaveOccurred())
				Expect(yml).To(HelpfullyMatchYAML(expectedYml))
			})
		})
	})
})
