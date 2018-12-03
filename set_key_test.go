package traverser_test

import (
	"github.com/mikesimons/traverser"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("traverser", func() {
	Describe("SetKey", func() {
		It("should set value at target", func() {
			data := mapTestData()

			err := traverser.SetKey(data, []string{"a", "aa"}, "def")
			val := data["a"].(map[interface{}]interface{})["aa"]

			Expect(err).To(BeNil())
			Expect(val).To(Equal("def"))
		})

		It("should return error if node is not traversable", func() {
			data := mapTestData()

			err := traverser.SetKey(data, []string{"a", "aa", "test"}, "value")
			Expect(err.Error()).To(Equal("can't traverse string at a.aa"))
		})
	})
})
