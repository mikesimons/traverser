package traverser

import (
	"fmt"
	"reflect"
	"strings"
)

// GetKey will return the value at the target key
func GetKey(data interface{}, target []string) (interface{}, error) {
	v, err := getKey(reflect.ValueOf(data), target, []string{})
	if err != nil {
		return nil, err
	}
	return v.Interface(), nil
}

func getKey(data reflect.Value, target []string, traversed []string) (reflect.Value, error) {
	nextKey := target[0]
	nextTarget := target[1:]
	var zeroVal reflect.Value

	switch data.Kind() {
	case reflect.Interface:
		return getKey(data.Elem(), target, traversed)
	case reflect.Ptr:
		return getKey(reflect.Indirect(data), target, traversed)
	case reflect.Map:
		nextKeyVal := reflect.ValueOf(nextKey)
		nextVal := data.MapIndex(nextKeyVal)

		if nextVal == zeroVal {
			traversedString := ""
			if len(traversed) > 0 {
				traversedString = fmt.Sprintf(" beyond %s", strings.Join(traversed, "."))
			}
			return reflect.Value{}, fmt.Errorf("key does not exist%s", traversedString)
		}

		if len(nextTarget) == 0 {
			return nextVal.Elem(), nil
		}

		traversed = append(traversed, nextKey)
		return getKey(nextVal, nextTarget, traversed)
	default:
		return reflect.Value{}, fmt.Errorf("can't traverse %s at %s", data.Kind().String(), strings.Join(traversed, "."))
	}
}
