package traverser

import (
	"fmt"
	"reflect"
	"strings"
)

// SetKey will set the value of the target key
func SetKey(data interface{}, target []string, value interface{}) error {
	return setKey(reflect.ValueOf(data), target, []string{}, reflect.ValueOf(value))
}

func setKey(data reflect.Value, target []string, traversed []string, value reflect.Value) error {
	nextKey := target[0]
	nextTarget := target[1:]
	var zeroVal reflect.Value

	switch data.Kind() {
	case reflect.Interface:
		return setKey(data.Elem(), target, traversed, value)
	case reflect.Ptr:
		return setKey(reflect.Indirect(data), target, traversed, value)
	case reflect.Map:
		nextKeyVal := reflect.ValueOf(nextKey)
		nextVal := data.MapIndex(nextKeyVal)

		if len(nextTarget) == 0 {
			data.SetMapIndex(nextKeyVal, value)
			return nil
		}

		if nextVal == zeroVal {
			nextVal = reflect.ValueOf(make(map[string]interface{}))
			data.SetMapIndex(nextKeyVal, nextVal)
		}

		traversed = append(traversed, nextKey)
		return setKey(nextVal, nextTarget, traversed, value)
	default:
		return fmt.Errorf("can't traverse %s at %s", data.Kind().String(), strings.Join(traversed, "."))
	}
}
