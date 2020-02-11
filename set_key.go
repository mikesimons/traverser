package traverser

import (
	"fmt"
	"reflect"
	"strings"
)

// SetKey will set the value of the target key
func SetKey(data interface{}, target []string, value interface{}) (reflect.Value, error) {
	var set bool
	t := &Traverser{
		Node: func(keys []string, val reflect.Value) (Op, error) {
			if reflect.DeepEqual(keys, target) {
				set = true
				return Set(reflect.ValueOf(value))
			}
			return Noop()
		},
	}

	ret, err := t.Traverse(reflect.ValueOf(&data))

	if !set {
		return ret, fmt.Errorf("Invalid field: %s", strings.Join(target, "."))
	}

	return ret, err
}
