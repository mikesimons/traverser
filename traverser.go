package traverser

import (
	"fmt"
	"reflect"
)

// New traversal code is derivative of https://gist.github.com/hvoecking/10772475
// MIT License; Copyright (c) 2014 Heye Vöcking - See LICENSE for full license

const (
	opNoop = iota
	opSet
	opUnset
	opSkip
	opSplice
)

// Traverser is the main type and contains the Map & Node callbacks to be used.
// Map will be called each time a Map type is encountered
// Node will be called for each traversable element
// Accept will be called each time a traversal is made
type Traverser struct {
	Map    func(keys []string, key string, data reflect.Value)
	Node   func(keys []string, data reflect.Value) (Op, error)
	Accept func(keys []string, data reflect.Value) (Op, error)
}

// Op represents an operation to perform on a value passed to the Node callback.
// It is used to skip, mutate and handle error conditions.
type Op struct {
	op  int
	val reflect.Value
}

// Traverse is the entrypoint for recursive traversal of the given reflect.Value.
// reflect.Value is used to support both maps and lists at the root as interface{} is not compatible with []interface{}
func (gt *Traverser) Traverse(original reflect.Value) (reflect.Value, error) {
	if !original.IsValid() {
		return reflect.Value{}, nil
	}

	copy := reflect.New(original.Type()).Elem()
	_, err := gt.traverse(copy, original, []string{})
	return copy, err
}

func (gt *Traverser) traverse(copy, original reflect.Value, keys []string) (Op, error) {
	if gt.Accept != nil && len(keys) > 0 {
		op, _ := gt.Accept(keys, copy)
		if op.op == opSkip {
			return Noop()
		}
	}

	switch original.Kind() {
	case reflect.Ptr:
		originalValue := original.Elem()
		if !originalValue.IsValid() {
			return Noop()
		}
		copy.Set(reflect.New(originalValue.Type()))
		return gt.traverse(copy.Elem(), originalValue, keys)

	case reflect.Interface:
		originalValue := original.Elem()
		if !originalValue.IsValid() {
			return Noop()
		}
		copyValue := reflect.New(originalValue.Type()).Elem()
		op, err := gt.traverse(copyValue, originalValue, keys)
		copy.Set(copyValue)
		return op, err

	case reflect.Struct:
		for i := 0; i < original.NumField(); i++ {
			if original.Field(i).CanSet() {
				op, err := gt.traverse(copy.Field(i), original.Field(i), append(keys, original.Type().Field(i).Name))
				if err != nil {
					return op, err
				}
			}
		}

		if gt.Node != nil {
			return gt.Node(keys, original)
		}

	case reflect.Slice:
		copy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))

		ci := -1
		for i := 0; i < original.Len(); i++ {
			ci++
			copyValue := copy.Interface().([]interface{})
			op, err := gt.traverse(copy.Index(ci), original.Index(i), append(keys, fmt.Sprintf("%d", i)))
			if err != nil {
				return op, err
			}

			if op.op == opSet {
				copyValue[i] = op.val.Interface()
				copy.Set(reflect.ValueOf(copyValue))
			} else if op.op == opUnset {
				copy.Set(reflect.ValueOf(append(copyValue[:ci], copyValue[ci+1:]...)))
				ci--
			} else if op.op == opSplice {
				splice, ok := op.val.Interface().([]interface{})
				if !ok {
					return ErrorNoop(fmt.Errorf("only slices can be spliced"))
				}
				copyValue = append(copyValue[:ci], append(splice, copyValue[ci+1:]...)...)
				copy.Set(reflect.ValueOf(copyValue))
				ci += len(splice) - 1
			}
		}

		if gt.Node != nil {
			return gt.Node(keys, original)
		}

	case reflect.Map:
		copy.Set(reflect.MakeMap(original.Type()))
		for _, key := range original.MapKeys() {
			originalValue := original.MapIndex(key)
			copyValue := reflect.New(originalValue.Type()).Elem()

			keyString := fmt.Sprintf("%v", key)
			if gt.Map != nil {
				gt.Map(keys, keyString, originalValue)
			}

			op, err := gt.traverse(copyValue, originalValue, append(keys, keyString))
			copy.SetMapIndex(key, copyValue)

			if err != nil {
				return op, err
			}

			if op.op == opSet || op.op == opUnset {
				copy.SetMapIndex(key, op.val)
			}
		}

		if gt.Node != nil {
			return gt.Node(keys, original)
		}

	default:
		copy.Set(original)
		if gt.Node != nil {
			return gt.Node(keys, original)
		}
	}

	return Noop()
}

// Set is a helper function that will return an Op to set the key currently being traversed to the given value
func Set(v reflect.Value) (Op, error) {
	return Op{opSet, v}, nil
}

// Noop is a helper function that will return an Op that doesn't do anything
func Noop() (Op, error) {
	return Op{opNoop, reflect.Value{}}, nil
}

// Unset is a helper function that will return an Op that unsets the key currently being traversed
func Unset() (Op, error) {
	return Op{opUnset, reflect.Value{}}, nil
}

func Splice(v reflect.Value) (Op, error) {
	return Op{opSplice, v}, nil
}

// ErrorSet is a helper function that will return an Op that sets the key currently being traversed to the given value and returns an error
func ErrorSet(err error, v reflect.Value) (Op, error) {
	return Op{opSet, v}, err
}

// ErrorUnset is a helper function that will return an Op that unsets the key currently being traversed and returns an error
func ErrorUnset(err error) (Op, error) {
	return Op{opUnset, reflect.Value{}}, err
}

// ErrorNoop is a helper function that will return an Op that doesn't do anything but return an error
func ErrorNoop(err error) (Op, error) {
	return Op{opNoop, reflect.Value{}}, err
}

// Skip is a helper function that will return an Op that will skip processing of the current node
func Skip() (Op, error) {
	return Op{opSkip, reflect.Value{}}, nil
}
