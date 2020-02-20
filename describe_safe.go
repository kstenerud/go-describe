// +build js appengine safe

package describe

import (
	"reflect"
)

func initUnsafe() {
}

func canExposeInterface() bool {
	return false
}

func exposeInterface(v reflect.Value) interface{} {
	return "go-describe.BUG(exposeInterface called from a safe build)"
}
