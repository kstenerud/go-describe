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

func initNestedReflectValues() {
	if field, ok := reflect.TypeOf(reflect.Value{}).FieldByName("flag"); ok {
		flagOffset = field.Offset
		hasExpectedReflectStruct = true
	} else {
		log.Printf("go-describe: Unsafe operations disabled because the " +
			"reflect.Value struct no longer has a flags field. Please open an " +
			"issue at https://github.com/kstenerud/go-describe\n")
		return
	}

	getReflectFlag := func(v reflect.Value) flag {
		return flag(reflect.ValueOf(v).FieldByName("flag").Uint())
	}
	rv := reflect.ValueOf(flagChecker{})
	roFlagMask = ^getReflectFlag(rv.FieldByName("a")) ^ getReflectFlag(rv.FieldByName("A"))
}

func canDereferenceNestedReflectValues() bool {
	return hasExpectedReflectStruct && EnableUnsafeOperations
}

func dereferenceNestedReflectValue(v reflect.Value) reflect.Value {
	// Note: This is the only unsafe operation.
	roFlag := (*flag)(unsafe.Pointer(uintptr(unsafe.Pointer(&v)) + flagOffset))
	*roFlag &= roFlagMask
	return v.Interface().(reflect.Value)
}
