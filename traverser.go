package traverser

import (
	"fmt"
	"reflect"
)

const op_noop = 0
const op_set = 1
const op_unset = 2

type Op struct {
	op  int
	val reflect.Value
}

type Traverser struct {
	Map  func(keys []string, key string, data reflect.Value)
	Node func(keys []string, data reflect.Value) (Op, error)
}

func (gt *Traverser) Traverse(data reflect.Value) error {
	_, err := gt.traverse(data, make([]string, 0))
	return err
}

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

func Set(v reflect.Value) (Op, error) {
	return Op{op_set, v}, nil
}

func Noop() (Op, error) {
	return Op{op_noop, reflect.Value{}}, nil
}

func Unset() (Op, error) {
	return Op{op_unset, reflect.Value{}}, nil
}

func ErrorUnset(err error) (Op, error) {
	return Op{op_unset, reflect.Value{}}, err
}

func ErrorNoop(err error) (Op, error) {
	return Op{op_noop, reflect.Value{}}, err
}
