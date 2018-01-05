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

type MapVisitor interface {
	Map(keys []string, key string, data reflect.Value)
}

type NodeVisitor interface {
	Node(keys []string, data reflect.Value) (Op, error)
}

type Traverser struct {
	Map  func(keys []string, key string, data reflect.Value)
	Node func(keys []string, data reflect.Value) (Op, error)
}

func New(r interface{}) *Traverser {
	ret := &Traverser{}
	if mapVisitor, ok := r.(MapVisitor); ok {
		ret.Map = mapVisitor.Map
	}

	if nodeVisitor, ok := r.(NodeVisitor); ok {
		ret.Node = nodeVisitor.Node
	}

	return ret
}

func (gt *Traverser) Traverse(data reflect.Value, keys []string) (Op, error) {
	if data.Kind() == reflect.Interface {
		data = data.Elem()
	}

	switch data.Kind() {
	case reflect.Map:
		for _, k := range data.MapKeys() {
			v := data.MapIndex(k)
			ks := fmt.Sprintf("%v", k)
			if gt.Map != nil {
				gt.Map(keys, ks, v)
			}

			op, _ := gt.Traverse(v, append(keys, ks))
			if op.op == op_set || op.op == op_unset {
				data.SetMapIndex(k, op.val)
			}
		}
	case reflect.Slice:
		d := data.Interface().([]interface{})
		for k := range d {
			lastKey := ""
			lastKeyIdx := len(keys) - 1
			if lastKeyIdx >= 0 {
				lastKey = keys[lastKeyIdx]
			}
			op, _ := gt.Traverse(reflect.ValueOf(d[k]), append(keys, fmt.Sprintf("%v%d", lastKey, k)))
			if op.op == op_set || op.op == op_unset {
				d[k] = op.val.Elem()
			}
		}
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

func Error(err error) (Op, error) {
	return Op{op_noop, reflect.Value{}}, err
}
