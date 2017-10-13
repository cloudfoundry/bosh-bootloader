package terraform_test

import (
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TerraformOutputs", func() {
	Describe("GetString", func() {
		It("returns the string value for the key", func() {
			outputs := terraform.Outputs{Map: map[string]interface{}{"foo": "bar"}}
			Expect(outputs.GetString("foo")).To(Equal("bar"))
		})

		Context("when the key does not exist", func() {
			It("returns an empty string", func() {
				outputs := terraform.Outputs{Map: map[string]interface{}{"foo": "bar"}}
				Expect(outputs.GetString("baz")).To(Equal(""))
			})
		})

		Context("when the value is not a string", func() {
			It("returns an empty string", func() {
				outputs := terraform.Outputs{Map: map[string]interface{}{"foo": []string{"bar", "baz"}}}
				Expect(outputs.GetString("foo")).To(Equal(""))
			})
		})
	})

	Describe("GetStringSlice", func() {
		Context("when the value is an interface slice", func() {
			It("returns the string slice value for the key", func() {
				outputs := terraform.Outputs{Map: map[string]interface{}{"foo": []interface{}{"bar", "baz"}}}
				Expect(outputs.GetStringSlice("foo")).To(ConsistOf([]string{"bar", "baz"}))
				Expect(outputs.GetStringSlice("foo")[0]).To(Equal("bar"))
				Expect(outputs.GetStringSlice("foo")[1]).To(Equal("baz"))
			})
		})

		Context("when the value is a string slice", func() {
			It("returns the string slice value for the key", func() {
				outputs := terraform.Outputs{Map: map[string]interface{}{"foo": []string{"bar", "baz"}}}
				Expect(outputs.GetStringSlice("foo")).To(ConsistOf([]string{"bar", "baz"}))
				Expect(outputs.GetStringSlice("foo")[0]).To(Equal("bar"))
				Expect(outputs.GetStringSlice("foo")[1]).To(Equal("baz"))
			})
		})

		Context("when the key does not exist", func() {
			It("returns an empty string slice", func() {
				outputs := terraform.Outputs{Map: map[string]interface{}{"foo": []interface{}{"bar", "baz"}}}
				Expect(outputs.GetStringSlice("baz")).To(BeEmpty())
				Expect(outputs.GetStringSlice("baz")).To(BeAssignableToTypeOf([]string{}))
			})
		})

		Context("when the value is not a slice", func() {
			It("returns an empty string slice", func() {
				outputs := terraform.Outputs{Map: map[string]interface{}{"foo": "bar"}}
				Expect(outputs.GetStringSlice("foo")).To(BeEmpty())
				Expect(outputs.GetStringSlice("foo")).To(BeAssignableToTypeOf([]string{}))
			})
		})

		Context("when the value is a slice of non-strings", func() {
			It("returns an empty string slice", func() {
				outputs := terraform.Outputs{Map: map[string]interface{}{"foo": []interface{}{1, 2, 3}}}
				Expect(outputs.GetStringSlice("foo")).To(BeEmpty())
				Expect(outputs.GetStringSlice("foo")).To(BeAssignableToTypeOf([]string{}))
			})
		})

		Context("when the value is a slice of mixed types", func() {
			It("returns an empty string slice", func() {
				outputs := terraform.Outputs{Map: map[string]interface{}{"foo": []interface{}{"bar", 1}}}
				Expect(outputs.GetStringSlice("foo")).To(BeEmpty())
				Expect(outputs.GetStringSlice("foo")).To(BeAssignableToTypeOf([]string{}))
			})
		})
	})

	Describe("GetStringMap", func() {
		var (
			mapFixture map[string]string
			emptyMap   map[string]string
		)

		BeforeEach(func() {
			mapFixture = map[string]string{"bar": "baz"}
			emptyMap = map[string]string{}
		})

		It("returns the string map value for the key", func() {
			outputs := terraform.Outputs{Map: map[string]interface{}{"foo": mapFixture}}
			Expect(outputs.GetStringMap("foo")).To(Equal(mapFixture))
		})

		Context("when the value is a string to interface map", func() {
			It("returns the string to string map value for the key", func() {
				outputs := terraform.Outputs{Map: map[string]interface{}{"foo": map[string]interface{}{"bar": "baz"}}}
				Expect(outputs.GetStringMap("foo")).To(Equal(mapFixture))
			})
		})

		Context("when the map value has a non-string value", func() {
			It("returns the string to string map value for the key", func() {
				outputs := terraform.Outputs{Map: map[string]interface{}{"foo": map[string]interface{}{"bar": 3}}}
				Expect(outputs.GetStringMap("foo")).To(Equal(emptyMap))
			})
		})

		Context("when the value is not a map", func() {
			It("returns an empty map", func() {
				outputs := terraform.Outputs{Map: map[string]interface{}{"foo": 1}}
				Expect(outputs.GetStringMap("foo")).To(Equal(emptyMap))
			})
		})

		Context("when the key is missing", func() {
			It("returns an empty map", func() {
				outputs := terraform.Outputs{Map: map[string]interface{}{}}
				Expect(outputs.GetStringMap("foo")).To(Equal(emptyMap))
			})
		})
	})
})
