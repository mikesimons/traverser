package traverser_test

import (
	"testing"

	"github.com/mikesimons/traverser"

	//"fmt"
	"reflect"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rounds/go-optikon"
)

func TestTraverser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Traverser Suite")
}

type testStruct struct {
	Value string
}

func mapTestData() map[interface{}]interface{} {
	return map[interface{}]interface{}{
		"a": map[interface{}]interface{}{
			"aa": "abc",
			"ab": 123,
			"ac": map[interface{}]interface{}{
				"aca": "hello",
			},
		},
		"b": []interface{}{
			"one",
			2,
			"three",
			nil,
		},
		"c": testStruct{Value: "hello"},
		"d": &testStruct{Value: "hello"},
	}
}

func noopNodeVisitor(keys []string, data reflect.Value) (traverser.Op, error) {
	return traverser.Noop()
}

var _ = Describe("traverser", func() {
	Describe("Traverse", func() {
		It("should call Node callback for each node", func() {
			nodes := make(map[string]interface{})
			// Make some of these interface and some *interface types. Test by validating key paths + values against expected list
			t := traverser.Traverser{
				Node: func(keys []string, data reflect.Value) (traverser.Op, error) {
					nodes[strings.Join(keys, "/")] = data.Interface()
					return traverser.Noop()
				},
			}

			data := mapTestData()
			err := t.Traverse(reflect.ValueOf(data))

			Expect(err).To(BeNil())
			Expect(len(nodes)).To(Equal(6))

			// spot checks
			Expect(nodes["a/ac/aca"]).To(Equal("hello"))
			Expect(nodes["a/ab"]).To(Equal(123))
			Expect(nodes["b/0"]).To(Equal("one"))
			Expect(nodes["b/1"]).To(Equal(2))
		})

		It("should call Map callback for each map", func() {
			maps := make(map[string]interface{})
			t := traverser.Traverser{
				Map: func(keys []string, key string, data reflect.Value) {
					maps[strings.Join(keys, "/")] = data.Interface()
				},
			}

			data := mapTestData()
			err := t.Traverse(reflect.ValueOf(data))

			Expect(err).To(BeNil())
			Expect(len(maps)).To(Equal(3))
			Expect(maps[""]).NotTo(BeNil())
			Expect(maps["a"]).NotTo(BeNil())
			Expect(maps["a/ac"]).NotTo(BeNil())
		})

		It("should handle nil reflect values", func() {
			t := traverser.Traverser{
				Node: noopNodeVisitor,
			}

			err := t.Traverse(reflect.ValueOf(nil))

			Expect(err).To(BeNil())
		})
	})

	Describe("Traverse return operation handling", func() {
		It("should change map value if OP_SET returned from Node", func() {
			t := traverser.Traverser{
				Node: func(keys []string, data reflect.Value) (traverser.Op, error) {
					if strings.Join(keys, "/") == "a/aa" {
						return traverser.Set(reflect.ValueOf("TEST"))
					}
					return traverser.Noop()
				},
			}

			data := mapTestData()
			err := t.Traverse(reflect.ValueOf(data))
			Expect(err).To(BeNil())

			splitPath := []string{"a", "aa"}
			v, err := optikon.Select(data, splitPath)
			Expect(err).To(BeNil())
			Expect(v).To(Equal("TEST"))
		})

		It("should unset map value if OP_UNSET returned from Node", func() {
			t := traverser.Traverser{
				Node: func(keys []string, data reflect.Value) (traverser.Op, error) {
					if strings.Join(keys, "/") == "a/aa" {
						return traverser.Unset()
					}
					return traverser.Noop()
				},
			}

			data := mapTestData()
			err := t.Traverse(reflect.ValueOf(data))
			Expect(err).To(BeNil())

			splitPath := []string{"a", "aa"}
			_, err = optikon.Select(data, splitPath)
			Expect(err).To(BeAssignableToTypeOf(&optikon.KeyNotFoundError{}))
		})

		It("should change slice value if OP_SET returned from Node", func() {
			t := traverser.Traverser{
				Node: func(keys []string, data reflect.Value) (traverser.Op, error) {
					if strings.Join(keys, "/") == "b/0" {
						return traverser.Set(reflect.ValueOf("TEST"))
					}
					return traverser.Noop()
				},
			}

			data := mapTestData()
			err := t.Traverse(reflect.ValueOf(data))
			Expect(err).To(BeNil())

			splitPath := []string{"b", "0"}
			v, err := optikon.Select(data, splitPath)
			Expect(err).To(BeNil())
			Expect(v).To(Equal("TEST"))
		})

		It("should unset slice value if OP_UNSET returned from Node", func() {
			t := traverser.Traverser{
				Node: func(keys []string, data reflect.Value) (traverser.Op, error) {
					if strings.Join(keys, "/") == "b/0" {
						return traverser.Unset()
					}
					return traverser.Noop()
				},
			}

			data := mapTestData()

			splitPath := []string{"b", "0"}
			oldval, _ := optikon.Select(data, splitPath)

			expected, _ := optikon.Select(data, []string{"b", "1"})

			err := t.Traverse(reflect.ValueOf(data))
			Expect(err).To(BeNil())

			newval, _ := optikon.Select(data, splitPath)
			Expect(newval).NotTo(Equal(oldval))
			Expect(newval).To(Equal(expected))
		})

		PIt("should do nothing to map if OP_NOOP returned from Node")
		PIt("should do nothing to slice if OP_NOOP returned from Node")
	})
})
