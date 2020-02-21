// +build go1.12

package describe

import (
	"reflect"
)

func mapRange(v reflect.Value) *reflect.MapIter {
	return v.MapRange()
}
