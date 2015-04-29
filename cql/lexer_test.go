package cql

import (
	"bufio"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cassandra CQL Lexer", func() {

	Context("Lexer", func() {

		// These test are actually really brittle since it only takes a space character
		// on the end of a line to break them.
		It("should return multiple cql statements", func() {
			exp := `select * from schema_version;select * from
my_table;`
			s := bufio.NewScanner(strings.NewReader(exp))
			s.Split(scanCQLExpressions)

			Expect(s.Scan()).To(Equal(true))
			Expect(s.Text()).To(Equal("select * from schema_version"))
			Expect(s.Scan()).To(Equal(true))
			Expect(s.Text()).To(Equal("select * from\nmy_table"))
			Expect(s.Scan()).To(Equal(false))
		})

		It("should ignore funny dash-dash SQL style comments", func() {
			exp := `select * from schema_version;
                    -- select * from yet_another-table
                    `
			s := bufio.NewScanner(strings.NewReader(exp))
			s.Split(scanCQLExpressions)

			Expect(s.Scan()).To(Equal(true))
			Expect(s.Text()).To(Equal("select * from schema_version"))
			Expect(s.Scan()).To(Equal(true))
			Expect(s.Text()).To(Equal("                    ")) // This nastiness is a temporary work around for lack of
			Expect(s.Scan()).To(Equal(false))                  // ability to handle a comment that ends with no newline after it.
		})

		// TODO: It's important to note that this test will hang if un-excluded (see above)
		XIt("should cope with comments that DON'T have a newline at the end", func() {
			exp := `select * from schema_version;
                    -- select * from yet_another-table`

			s := bufio.NewScanner(strings.NewReader(exp))
			s.Split(scanCQLExpressions)

			Expect(s.Scan()).To(Equal(true))
			Expect(s.Text()).To(Equal("select * from schema_version"))
			Expect(s.Scan()).To(Equal(false))
		})

		XIt("should crap itself on multiline comments", func() {
			Expect(scanCQLExpressions([]byte("/*"), false)).Should(Panic()) // Can you actually 'catch' a panic. I guess no...so what's this for and how do I use it?
		})

		It("should ignore C style comments", func() {
			exp := `select * from schema_version; // select * from
another_token;
    //select
* from
yet_another-table;`

			s := bufio.NewScanner(strings.NewReader(exp))
			s.Split(scanCQLExpressions)

			Expect(s.Scan()).To(Equal(true))
			Expect(s.Text()).To(Equal("select * from schema_version"))
			Expect(s.Scan()).To(Equal(true))
			Expect(s.Text()).To(Equal("another_token"))
			Expect(s.Scan()).To(Equal(true))
			Expect(s.Text()).To(Equal("* from\nyet_another-table"))
			Expect(s.Scan()).To(Equal(false))
		})
	})
})
