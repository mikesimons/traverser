package traverser_test

import (
	"github.com/mikesimons/traverser"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("traverser", func() {
	Describe("GetKey", func() {
		It("should get value at target", func() {
			data := mapTestData()

			val, err := traverser.GetKey(data, []string{"a", "aa"})
			Expect(err).To(BeNil())
			Expect(val).To(Equal("abc"))
		})

		It("should return error if target does not exist", func() {
			data := mapTestData()

			val, err := traverser.GetKey(data, []string{"zzzzz"})
			Expect(val).To(BeNil())
			Expect(err.Error()).To(Equal("key does not exist"))
		})

		It("should return error if node is not traversable", func() {
			data := mapTestData()

			val, err := traverser.GetKey(data, []string{"a", "aa", "test"})
			Expect(val).To(BeNil())
			Expect(err.Error()).To(Equal("can't traverse string at a.aa"))
		})
	})
})
