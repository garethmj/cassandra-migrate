package cql

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cassandra Migrations", func() {

	Context("Multiple errors returned from a func", func() {
		It("should return multiple errors", func() {
			var errs Errors
			errs = append(errs, fmt.Errorf("My first error"))
			errs = append(errs, fmt.Errorf("My second error"))

			Expect(len(errs)).To(Equal(2))
			Expect(errs.Error()).To(Equal("Multiple Errors:\n  My first error\n  My second error"))
		})
	})
})
