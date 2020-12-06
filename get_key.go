package traverser

import (
	"fmt"
	"reflect"
	"strings"
)

// GetKey will return the value at the target key
func GetKey(data interface{}, target []string) (interface{}, error) {
	var found bool
	var output interface{}
	t := &Traverser{
		Node: func(keys []string, val reflect.Value) (Op, error) {
			if reflect.DeepEqual(keys, target) {
				found = true
				output = val.Interface()
			}
			return Noop()
		},
	}

	t.Traverse(reflect.ValueOf(&data))

	if !found {
		return output, fmt.Errorf("Invalid field: %s", strings.Join(target, "."))
	}

	return output, nil
}
