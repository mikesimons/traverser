// Package traverser implements traversal of unknown structures with optional callbacks
package traverser

import (
	"fmt"
	"reflect"
)

const (
	op_noop = iota
	op_set
	op_unset
)

// Traverser is the main type and contains the Map & Node callbacks to be used.
// Map will be called each time a Map type is encountered
// Node will be called for each traversable element
type Traverser struct {
	Map  func(keys []string, key string, data reflect.Value)
	Node func(keys []string, data reflect.Value) (Op, error)
}

// Op represents an operation to perform on a value passed to the Node callback.
// It is used to skip, mutate and handle error conditions.
type Op struct {
	op  int
	val reflect.Value
}

// Traverse is the recursive entrypoint for traversal of the given reflect.Value.
func (gt *Traverser) Traverse(data reflect.Value) error {
	_, err := gt.traverse(data, make([]string, 0))
	return err
}

// traverse is the internal recursion function and handles the core traversal logic.
func (gt *Traverser) traverse(data reflect.Value, keys []string) (Op, error) {
	switch data.Kind() {
	case reflect.Interface:
		return gt.traverse(data.Elem(), keys)
	case reflect.Ptr:
		return gt.traverse(reflect.Indirect(data), keys)
	case reflect.Map:
		for _, k := range data.MapKeys() {
			v := data.MapIndex(k)
			ks := fmt.Sprintf("%v", k)
			if gt.Map != nil {
				gt.Map(keys, ks, v)
			}

			op, err := gt.traverse(v, append(keys, ks))
			if err != nil {
				return op, err
			}

			if op.op == op_set || op.op == op_unset {
				data.SetMapIndex(k, op.val)
			}
		}
	case reflect.Slice:
		d := data.Interface().([]interface{})
		for k := range d {
			if k >= len(d) {
				return Noop()
			}

			op, err := gt.traverse(reflect.ValueOf(d[k]), append(keys, fmt.Sprintf("%v", k)))
			if err != nil {
				return op, err
			}

			if op.op == op_set {
				d[k] = op.val.Interface()
			} else if op.op == op_unset {
				d = append(d[:k], d[k+1:]...)
			}
		}
	case reflect.Struct:
		fallthrough
	case reflect.Invalid:
		return Noop()
	default:
		if gt.Node != nil {
			return gt.Node(keys, data)
		}
	}

	return Noop()
}

// Set is a helper function that will return an Op to set the key currently being traversed to the given value
func Set(v reflect.Value) (Op, error) {
	return Op{op_set, v}, nil
}

// Noop is a helper function that will return an Op that doesn't do anything
func Noop() (Op, error) {
	return Op{op_noop, reflect.Value{}}, nil
}

// Unset is a helper function that will return an Op that unsets the key currently being traversed
func Unset() (Op, error) {
	return Op{op_unset, reflect.Value{}}, nil
}

// ErrorSet is a helper function that will return an Op that sets the key currently being traversed to the given value and returns an error
func ErrorSet(err error, v reflect.Value) (Op, error) {
	return Op{op_set, v}, err
}

// ErrorUnset is a helper function that will return an Op that unsets the key currently being traversed and returns an error
func ErrorUnset(err error) (Op, error) {
	return Op{op_unset, reflect.Value{}}, err
}

// ErrorNoop is a helper function that will return an Op that doesn't do anything but return an error
func ErrorNoop(err error) (Op, error) {
	return Op{op_noop, reflect.Value{}}, err
}
