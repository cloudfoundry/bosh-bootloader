package multierror_test

import (
	"fmt"

	"code.cloudfoundry.org/multierror"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Multierror", func() {
	var (
		m *multierror.MultiError
	)

	BeforeEach(func() {
		m = multierror.NewMultiError("top level field")
	})

	Describe("Add", func() {
		Describe("adding a regular (non-MultiError) error", func() {
			It("creates a multierror with an empty list, and adds the error", func() {
				err := fmt.Errorf("Sample Error")
				m.Add(err)
				Expect(m.Length()).To(Equal(1))
				Expect(m.Error()).To(ContainSubstring("Sample Error"))
			})
		})

		Describe("adding a MultiError", func() {
			It("adds all the errors", func() {
				m.Add(fmt.Errorf("Error 1"))
				m.Add(fmt.Errorf("Error 2"))
				m2 := multierror.NewMultiError("inner field")
				m2.Add(fmt.Errorf("Error 3"))
				m2.Add(fmt.Errorf("Error 4"))
				m.Add(m2)

				Expect(m.Errors).To(ContainElement(m2))
				Expect(m.Length()).To(Equal(4))
			})
		})
	})

	Describe("Length", func() {
		Context("when the MultiError is a leaf node", func() {
			BeforeEach(func() {
				m.Add(fmt.Errorf("leaf error"))
			})

			It("returns 1", func() {
				innerError := m.Errors[0]
				Expect(innerError.Length()).To(Equal(1))
			})
		})

		Context("when the MultiError is not a leaf node, but is empty", func() {
			It("returns 0", func() {
				Expect(m.Length()).To(Equal(0))
			})
		})

		Context("when there are nested empty MultiErrors", func() {
			BeforeEach(func() {
				innerMostError := multierror.NewMultiError("innermost field")
				innerError := multierror.NewMultiError("inner field")
				innerError.Add(innerMostError)
				m.Add(innerError)
			})

			It("returns 0", func() {
				Expect(m.Length()).To(Equal(0))
			})
		})

		Context("when the multi error contains sub-errors", func() {
			BeforeEach(func() {
				m.Add(fmt.Errorf("error 1"))
				m.Add(fmt.Errorf("error 2"))
				m.Add(fmt.Errorf("error 3"))
			})

			It("returns the number of sub-errors", func() {
				Expect(m.Length()).To(Equal(3))
			})

			Context("when the sub-errors are MultiErrors", func() {
				BeforeEach(func() {
					nestedError := multierror.NewMultiError("inner field")
					nestedError.Add(fmt.Errorf("inner error 1"))
					nestedError.Add(fmt.Errorf("inner error 2"))
					m.Add(nestedError)
				})

				It("includes the nested sub-errors in the length", func() {
					Expect(m.Length()).To(Equal(5))
				})
			})
		})
	})

	Describe("Error", func() {
		Context("when multierrors are empty", func() {
			It("should say there were 0 errors", func() {
				Expect(m.Error()).To(ContainSubstring("there were 0 errors"))
			})

			Context("when there are nested multierrors but no actual errors", func() {
				BeforeEach(func() {
					m.Add(multierror.NewMultiError("another nested field"))
				})
				It("should say there were 0 errors", func() {
					Expect(m.Error()).To(ContainSubstring("there were 0 errors"))
				})
			})
		})

		Context("when there are errors", func() {

			Context("when the MultiError is a leaf node", func() {
				BeforeEach(func() {
					m.Add(fmt.Errorf("leaf error"))
				})

				It("returns a header with length plus the error message", func() {
					innerError := m.Errors[0]
					Expect(m.Error()).To(ContainSubstring("there was 1 error"))
					Expect(m.Error()).To(ContainSubstring(innerError.Message))
				})
			})

			Context("when there are nested errors", func() {
				BeforeEach(func() {
					cfErrors := multierror.NewMultiError("cf")
					cfErrors.Add(fmt.Errorf("the file was blah"))
					cfErrors.Add(fmt.Errorf("the thing was foo"))

					stemcellErrors := multierror.NewMultiError("stemcell")
					stemcellErrors.Add(fmt.Errorf("it stole my money"))

					stubsErrors := multierror.NewMultiError("stubs")
					stubsErrors.Add(fmt.Errorf("top level error for stubs occurred"))

					stub1Errors := multierror.NewMultiError("stub \"my-stub-1\"")
					stub1Errors.Add(fmt.Errorf("it never puts the toilet seat down"))
					stub1Errors.Add(fmt.Errorf("it can't\n handle\n all these lines"))

					stub2Errors := multierror.NewMultiError("stub \"my-stub-2\"")
					stub2Errors.Add(fmt.Errorf("error 1"))
					stub2Errors.Add(fmt.Errorf("error 2"))

					stubsErrors.Add(stub1Errors)
					stubsErrors.Add(stub2Errors)

					m.Add(cfErrors)
					m.Add(stemcellErrors)
					m.Add(stubsErrors)
				})

				It("presents the nested-ness of the errors", func() {
					Expect(m.Error()).To(ContainSubstring(`there were 8 errors with 'top level field'`))
					Expect(m.Error()).To(ContainSubstring(`    there were 2 errors with 'cf':`))
					Expect(m.Error()).To(ContainSubstring(`        * the file was blah`))
					Expect(m.Error()).To(ContainSubstring(`        * the thing was foo`))
					Expect(m.Error()).To(ContainSubstring(`    there was 1 error with 'stemcell':`))
					Expect(m.Error()).To(ContainSubstring(`        * it stole my money`))
					Expect(m.Error()).To(ContainSubstring(`    there were 5 errors with 'stubs':`))
					Expect(m.Error()).To(ContainSubstring(`        * top level error for stubs occurred`))
					Expect(m.Error()).To(ContainSubstring(`        there were 2 errors with 'stub "my-stub-1"':`))
					Expect(m.Error()).To(ContainSubstring(`            * it never puts the toilet seat down`))
					Expect(m.Error()).To(ContainSubstring(`            * it can't`))
					Expect(m.Error()).To(ContainSubstring(`              handle`))
					Expect(m.Error()).To(ContainSubstring(`              all these lines`))
					Expect(m.Error()).To(ContainSubstring(`        there were 2 errors with 'stub "my-stub-2"':`))
					Expect(m.Error()).To(ContainSubstring(`            * error 1`))
					Expect(m.Error()).To(ContainSubstring(`            * error 2`))
				})
			})
		})
	})
})
