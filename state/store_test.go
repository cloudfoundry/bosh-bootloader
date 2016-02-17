package state_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pivotal-cf-experimental/bosh-bootloader/state"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Store", func() {
	var (
		store state.Store
		dir   string
	)

	BeforeEach(func() {
		var err error
		dir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		store = state.NewStore()
	})

	Describe("Merge", func() {
		It("stores the given sub-map in a state.json file located in the given directory", func() {
			err := store.Merge(dir, map[string]interface{}{
				"key": "value",
			})
			Expect(err).NotTo(HaveOccurred())

			contents, err := ioutil.ReadFile(filepath.Join(dir, "state.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).To(MatchJSON(`{
				"key": "value"
			}`))
		})

		It("merges the keys with any existing state", func() {
			err := store.Merge(dir, map[string]interface{}{
				"key": "value",
			})
			Expect(err).NotTo(HaveOccurred())

			err = store.Merge(dir, map[string]interface{}{
				"another-key": "another-value",
			})
			Expect(err).NotTo(HaveOccurred())

			contents, err := ioutil.ReadFile(filepath.Join(dir, "state.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).To(MatchJSON(`{
				"key": "value",
				"another-key": "another-value"
			}`))
		})

		Context("failure cases", func() {
			Context("when the state file has a permissions issue", func() {
				It("returns an error", func() {
					err := ioutil.WriteFile(filepath.Join(dir, "state.json"), []byte("{}"), os.ModePerm)
					Expect(err).NotTo(HaveOccurred())

					err = os.Chmod(dir, 0000)
					Expect(err).NotTo(HaveOccurred())

					err = store.Merge(dir, map[string]interface{}{
						"key": "value",
					})
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			Context("when the state file cannot be read", func() {
				It("returns an error", func() {
					err := ioutil.WriteFile(filepath.Join(dir, "state.json"), []byte("{}"), 0000)
					Expect(err).NotTo(HaveOccurred())

					err = store.Merge(dir, map[string]interface{}{
						"key": "value",
					})
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			Context("when the state file contains malformed JSON", func() {
				It("returns an error", func() {
					err := ioutil.WriteFile(filepath.Join(dir, "state.json"), []byte("%%%%%"), os.ModePerm)
					Expect(err).NotTo(HaveOccurred())

					err = store.Merge(dir, map[string]interface{}{
						"key": "value",
					})
					Expect(err).To(MatchError(ContainSubstring("invalid character")))
				})
			})

			Context("when the new map contains a key that cannot be marshaled to JSON", func() {
				It("returns an error", func() {
					err := store.Merge(dir, map[string]interface{}{
						"key": func() {},
					})
					Expect(err).To(MatchError(ContainSubstring("unsupported type: func()")))
				})
			})

			Context("when the state file cannot be written", func() {
				It("returns an error", func() {
					err := ioutil.WriteFile(filepath.Join(dir, "state.json"), []byte("{}"), 0444)
					Expect(err).NotTo(HaveOccurred())

					err = store.Merge(dir, map[string]interface{}{
						"key": "value",
					})
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})
	})

	Describe("GetString", func() {
		It("fetches the value for the given key", func() {
			err := ioutil.WriteFile(filepath.Join(dir, "state.json"), []byte(`{
				"key": "value"
			}`), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			value, ok, err := store.GetString(dir, "key")
			Expect(err).NotTo(HaveOccurred())
			Expect(ok).To(BeTrue())
			Expect(value).To(Equal("value"))
		})

		Context("failure cases", func() {
			Context("when the state file cannot be read", func() {
				It("returns an error", func() {
					err := ioutil.WriteFile(filepath.Join(dir, "state.json"), []byte("%%%%%"), os.ModePerm)
					Expect(err).NotTo(HaveOccurred())

					_, ok, err := store.GetString(dir, "key")
					Expect(err).To(MatchError(ContainSubstring("invalid character")))
					Expect(ok).To(BeFalse())
				})
			})

			Context("when the value is not a string", func() {
				It("returns an error", func() {
					err := ioutil.WriteFile(filepath.Join(dir, "state.json"), []byte(`{
						"key": 7
					}`), os.ModePerm)
					Expect(err).NotTo(HaveOccurred())

					_, ok, err := store.GetString(dir, "key")
					Expect(err).To(MatchError(ContainSubstring(`value at key "key" is not type "string"`)))
					Expect(ok).To(BeFalse())
				})
			})

			Context("when the key is not present", func() {
				It("returns not ok", func() {
					_, ok, err := store.GetString(dir, "key")
					Expect(err).NotTo(HaveOccurred())
					Expect(ok).To(BeFalse())
				})
			})
		})
	})
})
