package cql

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sort"
)

var _ = Describe("Cassandra Migrations", func() {

	Context("Sorting Migration objects", func() {

		var m1, m2, m3, m4 *Migration
		var mlist Migrations

		BeforeEach(func() {
			m1 = &Migration{
				Environment: "all",
				Name:        "my_migrations_1",
				Version:     "201408210600",
			}
			m2 = &Migration{
				Environment: "all",
				Name:        "my_migrations_2",
				Version:     "201508210600",
			}
			m3 = &Migration{
				Environment: "all",
				Name:        "my_migrations_3",
				Version:     "201408220600",
			}
			m4 = &Migration{
				Environment: "all",
				Name:        "my_migrations_4",
				Version:     "201408210601",
			}

			mlist = append(mlist, m1)
			mlist = append(mlist, m3)
			mlist = append(mlist, m2)
			mlist = append(mlist, m4)
		})
		It("should order a Migrations list in ascending value of the Version field", func() {
			Expect(mlist[0].Name).To(Equal("my_migrations_1"))
			Expect(mlist[1].Name).To(Equal("my_migrations_3"))

			sort.Sort(mlist)

			Expect(mlist[0].Name).To(Equal("my_migrations_1"))
			Expect(mlist[1].Name).To(Equal("my_migrations_4"))
		})
		It("should respond true when it contains a matching Migration", func() {

			c1 := &Migration{ // the same as m4, above
				Environment: "all",
				Name:        "my_migrations_4",
				Version:     "201408210601",
			}
			c2 := &Migration{
				Environment: "all",
				Name:        "my_migrations_not_present",
				Version:     "201408210601",
			}
			Expect(mlist.Contains(c1)).To(BeTrue())
			Expect(mlist.Contains(c2)).To(BeFalse())
		})
	})
})
