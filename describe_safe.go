// +build js appengine safe

package describe

import (
	"reflect"
)

func initInterfaceUnexported() {
}

func canInterfaceUnexported() bool {
	return false
}

func interfaceUnexported(v reflect.Value) interface{} {
	return "go-describe BUG: interfaceUnexported called from safe build"
}
