// +build js appengine safe

package describe

import (
	"reflect"
)

func initNestedReflectValues() {
}

func canDereferenceNestedReflectValues() bool {
	return false
}

func dereferenceNestedReflectValue(v reflect.Value) reflect.Value {
	return reflect.Zero(reflect.TypeOf(0))
}
