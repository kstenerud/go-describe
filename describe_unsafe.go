// Only enable unsafe operations if we're not compiling for GopherJS or App
// Engine, and we weren't built with '-tags safe'
// +build !js,!appengine,!safe

package describe

import (
	"log"
	"reflect"
	"unsafe"
)

type flag uintptr // reflect/value.go:flag

type flagROTester struct {
	A   int
	a   int // reflect/value.go:flagStickyRO
	int     // reflect/value.go:flagEmbedRO
	// Note: flagRO = flagStickyRO | flagEmbedRO
}

var flagOffset uintptr
var maskFlagRO flag
var hasExpectedReflectStruct bool

func initUnsafe() {
	if field, ok := reflect.TypeOf(reflect.Value{}).FieldByName("flag"); ok {
		flagOffset = field.Offset
	} else {
		log.Println("go-describe: exposeInterface() is disabled because the " +
			"reflect.Value struct no longer has a flag field. Please open an " +
			"issue at https://github.com/kstenerud/go-describe/issues")
		hasExpectedReflectStruct = false
		return
	}

	rv := reflect.ValueOf(flagROTester{})
	getFlag := func(v reflect.Value, name string) flag {
		return flag(reflect.ValueOf(v.FieldByName(name)).FieldByName("flag").Uint())
	}
	flagRO := (getFlag(rv, "a") | getFlag(rv, "int")) ^ getFlag(rv, "A")
	maskFlagRO = ^flagRO

	if flagRO == 0 {
		log.Println("go-describe: exposeInterface() is disabled because the " +
			"reflect flag type no longer has a flagEmbedRO or flagStickyRO bit. " +
			"Please open an issue at https://github.com/kstenerud/go-describe/issues")
		hasExpectedReflectStruct = false
		return
	}

	hasExpectedReflectStruct = true
}

func canExposeInterface() bool {
	return hasExpectedReflectStruct && EnableUnsafeOperations
}

// Expose an interface to an unexported field, subverting the type system.
//
// This is ONLY meant for inspecting unexported reflect.Value and reflect.Type
// fields in order to compensate for an oversight in the reflect API design.
//
// Do not use this power for evil; it will consume you and everything you love.
func exposeInterface(v reflect.Value) interface{} {
	pFlag := (*flag)(unsafe.Pointer(uintptr(unsafe.Pointer(&v)) + flagOffset))
	*pFlag &= maskFlagRO
	return v.Interface()
}
