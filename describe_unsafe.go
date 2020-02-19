// Only enable unsafe operations if we're not compiling for GopherJS or App
// Engine, and we weren't built with '-tags safe'
// +build !js,!appengine,!safe

package describe

import (
	"log"
	"reflect"
	"unsafe"
)

type flagChecker struct {
	a int
	A int
}

type flag uintptr

var roFlagMask flag
var flagOffset uintptr
var hasExpectedReflectStruct bool

func initInterfaceUnexported() {
	if field, ok := reflect.TypeOf(reflect.Value{}).FieldByName("flag"); ok {
		flagOffset = field.Offset
		hasExpectedReflectStruct = true
	} else {
		log.Printf("go-describe: Unsafe operations disabled because the " +
			"reflect.Value struct no longer has a flags field. Please open an " +
			"issue at https://github.com/kstenerud/go-describe\n")
		return
	}

	getFlag := func(v reflect.Value) flag {
		return flag(reflect.ValueOf(v).FieldByName("flag").Uint())
	}
	rv := reflect.ValueOf(flagChecker{})
	roFlagMask = ^getFlag(rv.FieldByName("a")) ^ getFlag(rv.FieldByName("A"))
}

func canInterfaceUnexported() bool {
	return hasExpectedReflectStruct && EnableUnsafeOperations
}

func interfaceUnexported(v reflect.Value) interface{} {
	// Note: This is the only unsafe operation.
	roFlag := (*flag)(unsafe.Pointer(uintptr(unsafe.Pointer(&v)) + flagOffset))
	*roFlag &= roFlagMask
	return v.Interface()
}
