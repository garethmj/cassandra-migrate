package cql

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("Cassandra Migrations", func() {

	Context("Creating migration objects", func() {
		It("should clean migration name of whitespace and unicode", func() {
			m := CreateMigration("my Êxtra messed up\tÑame", "all")
			Expect(m.Name).To(Equal("my_e_xtra_messed_up_n_ame"))
		})
	})
	Context("With migration objects", func() {

		var m1, m2, m3, m4 *Migration

		BeforeEach(func() {
			m1 = &Migration{
				Environment: "all",
				Name:        "my_migrations",
				Version:     time.Now().Format(migrationTimeFormat),
			}
			m2 = &Migration{
				Environment: "all",
				Name:        "my_migrations",
				Version:     time.Date(2013, 01, 01, 01, 00, 00, 00, time.UTC).Format(migrationTimeFormat),
			}
			m3 = &Migration{
				Environment: "all",
				Name:        "my_new_migration",
				Version:     time.Now().Format(migrationTimeFormat),
			}
			m4 = &Migration{
				Environment: "all",
				Name:        "my_migrations",
				Version:     time.Now().Format(migrationTimeFormat),
			}
		})

		It("should return false when migrations are not equal (version)", func() {
			Expect(m1.Compare(m2)).To(BeFalse())
		})
		It("should return false when migrations are not equal (name)", func() {
			Expect(m1.Compare(m3)).To(BeFalse())
		})
		It("should return true when migrations are equal", func() {
			Expect(m1.Compare(m4)).To(BeTrue())
		})
	})

	Context("With migration files", func() {
		var (
			updates Migrations
			errs    Errors
		)

		BeforeEach(func() {
			updates, errs = ListMigrationFiles("../migrations/test")
			Expect(errs).To(BeNil()) // Dammit - with Errors we now can't use HaveOccurred() matcher.
		})

		It("should get a list of migrations from files in a directory", func() {

			Expect(len(updates)).To(Equal(4))
			Expect(errs).To(BeNil())
			Expect(len(errs)).To(Equal(0))
			Expect(errs.Error()).To(Equal(""))

			// TODO: Actually check something meaningful here.
			for _, u := range updates {
				fmt.Printf("Migration: '%s'\n   Env: %s\n   Name: %s\n   Version: %s\n   Sum: %x\n   Applied On: %s\n",
					u.File,
					u.Environment,
					u.Name,
					u.Version,
					u.Sum,
					u.Applied)
			}
		})
		It("should know when an []*Migration contains a particular Migration", func() {

			m1 := &Migration{
				Environment: "all",
				Name:        "portal_init",
				Version:     time.Date(2014, 8, 21, 06, 00, 00, 00, time.UTC).Format(migrationTimeFormat),
			}
			m2 := &Migration{
				Environment: "all",
				Name:        "no_such_migration",
				Version:     time.Date(2014, 8, 21, 06, 00, 00, 00, time.UTC).Format(migrationTimeFormat),
			}
			m3 := &Migration{
				Environment: "differentenv",
				Name:        "portal_init",
				Version:     time.Date(2014, 8, 21, 06, 00, 00, 00, time.UTC).Format(migrationTimeFormat),
			}

			Expect(updates.Contains(m1)).To(BeTrue())
			Expect(updates.Contains(m2)).To(BeFalse())
			Expect(updates.Contains(m3)).To(BeFalse())
		})
		It("must return false when []*Migration is empty", func() {

			emptyUpdates := make(Migrations, 0)
			m1 := &Migration{
				Name: "portal_init",
			}
			Expect(emptyUpdates.Contains(m1)).To(BeFalse())
		})
	})
})
