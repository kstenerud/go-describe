// A go module for describing objects as a single line of text. It provides MUCH
// more information than the `%v` formatter does, allowing you to see more about
// complex objects at a glance.
//
// It handles recursive data, and can describe structures that would cause `%v`
// to stack overflow.
//
// The description is built as a single line, and is structured as follows:
//
//  * Basic types are printed as-is (as if printed via `%v`)
//  * Strings are enclosed in quotes `""`
//  * Nil pointers are simply printed as `nil`
//  * Pointers are preceded by `&`
//  * Interfaces are preceded by `@`
//  * Slices and arrays are enclosed in `[]`
//  * Slices and arrays of unsigned int types are printed as hex
//  * Maps are enclosed in `{}`, listing keys and values separated by `:`
//  * Structs are preceded by the type name, with fields enclosed in `()`, listing
//    field names and values separated by `:`
//  * Custom describers by convention print a type name, then a description
//    within `<>` (example: `url<http://example.com>`)
//  * Duplicate and cyclic references will be marked as follows:
//    - The first instance will be prepended by an ID and `=`
//    - Further instances will be replaced by a reference: `$` and the ID
//
// Note: describe will use the unsafe package to look inside nested reflect.Value
//       objects unless compiled with `-tags safe`, or if EnableUnsafeOperations
//       is set to false. It will also be disabled if compiling for GopherJS or
//       AppEngine.
package describe

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"
)

func init() {
	initNestedReflectValues()
}

const (
	openString         = `"`
	closeString        = `"`
	openArray          = "["
	closeArray         = "]"
	openMap            = "{"
	closeMap           = "}"
	openStruct         = "("
	closeStruct        = ")"
	itemSeparator      = " "
	keyValueSeparator  = ":"
	pointerPrefix      = "&"
	interfacePrefix    = "@"
	referenceSeparator = "="
	referencePrefix    = "$"
	nilPointer         = "nil"
)

var reflectValueType = reflect.ValueOf(reflect.ValueOf(true)).Type()
var reflectTypeType = reflect.TypeOf(reflect.TypeOf(true))

// -----------------
// Custom describers
// -----------------

var customDescribers = map[reflect.Type]CustomDescriber{
	reflect.TypeOf(url.URL{}):   describeURL,
	reflect.TypeOf(time.Time{}): describeTime,
}

func describeURL(v reflect.Value) string {
	realValue := v.Interface().(url.URL)
	return fmt.Sprintf(`url<%v>`, realValue.String())
}

func describeTime(v reflect.Value) string {
	realValue := v.Interface().(time.Time)
	return fmt.Sprintf(`time<%v>`, realValue.String())
}

// -----------------
// Duplicates finder
// -----------------

func findDuplicatesInternal(seenPointerCounts map[uintptr]int, v reflect.Value) {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		findDuplicatesInternal(seenPointerCounts, v.Elem())
	case reflect.Array, reflect.Slice:
		ptr := v.Pointer()
		if count, ok := seenPointerCounts[ptr]; ok {
			seenPointerCounts[ptr] = count + 1
			return
		}
		seenPointerCounts[ptr] = 1
		for i := 0; i < v.Len(); i++ {
			findDuplicatesInternal(seenPointerCounts, v.Index(i))
		}
	case reflect.Map:
		ptr := v.Pointer()
		if count, ok := seenPointerCounts[ptr]; ok {
			seenPointerCounts[ptr] = count + 1
			return
		}
		seenPointerCounts[ptr] = 1
		for iter := v.MapRange(); iter.Next(); {
			findDuplicatesInternal(seenPointerCounts, iter.Value())
		}
	case reflect.Struct:
		if v.CanAddr() {
			ptr := v.Addr().Pointer()
			if count, ok := seenPointerCounts[ptr]; ok {
				seenPointerCounts[ptr] = count + 1
				return
			}
			seenPointerCounts[ptr] = 1
		}
		for i := 0; i < v.NumField(); i++ {
			findDuplicatesInternal(seenPointerCounts, v.Field(i))
		}
	}
}

func findDuplicates(v reflect.Value) map[uintptr]int {
	seenPointerCounts := map[uintptr]int{}
	findDuplicatesInternal(seenPointerCounts, v)
	referenceName := 1
	referenceNames := map[uintptr]int{}
	for pointer, count := range seenPointerCounts {
		if count > 1 {
			referenceNames[pointer] = referenceName
			referenceName++
		}
	}
	return referenceNames
}

// -------------------
// Description context
// -------------------

type descriptionContext struct {
	stringBuilder  strings.Builder
	referenceNames map[uintptr]int
	seenReferences map[uintptr]bool
}

func (this *descriptionContext) describeArray(v reflect.Value) {
	asHex := false
	switch v.Type().Elem().Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		asHex = true
	}
	isFirst := true
	this.stringBuilder.WriteString(openArray)
	for i := 0; i < v.Len(); i++ {
		if isFirst {
			isFirst = false
		} else {
			this.stringBuilder.WriteString(itemSeparator)
		}
		this.describeReflect(v.Index(i), asHex)
	}
	this.stringBuilder.WriteString(closeArray)
}

func (this *descriptionContext) describeMap(v reflect.Value) {
	isFirst := true
	this.stringBuilder.WriteString(openMap)
	iter := v.MapRange()
	for iter.Next() {
		if isFirst {
			isFirst = false
		} else {
			this.stringBuilder.WriteString(itemSeparator)
		}
		this.describeReflect(iter.Key(), false)
		this.stringBuilder.WriteString(keyValueSeparator)
		this.describeReflect(iter.Value(), false)
	}
	this.stringBuilder.WriteString(closeMap)
}

func (this *descriptionContext) describeStruct(v reflect.Value) {
	isFirst := true
	this.stringBuilder.WriteString(v.Type().Name())
	this.stringBuilder.WriteString(openStruct)
	for i := 0; i < v.NumField(); i++ {
		if isFirst {
			isFirst = false
		} else {
			this.stringBuilder.WriteString(itemSeparator)
		}
		this.stringBuilder.WriteString(v.Type().Field(i).Name)
		this.stringBuilder.WriteString(keyValueSeparator)
		this.describeReflect(v.Field(i), false)
	}
	this.stringBuilder.WriteString(closeStruct)
}

func (this *descriptionContext) describeUint8(v uint8, asHex bool) {
	if asHex {
		this.stringBuilder.WriteString(fmt.Sprintf("0x%02x", v))
	} else {
		this.stringBuilder.WriteString(fmt.Sprintf("%v", v))
	}
}

func (this *descriptionContext) describeUint16(v uint16, asHex bool) {
	if asHex {
		this.stringBuilder.WriteString(fmt.Sprintf("0x%04x", v))
	} else {
		this.stringBuilder.WriteString(fmt.Sprintf("%v", v))
	}
}

func (this *descriptionContext) describeUint32(v uint32, asHex bool) {
	if asHex {
		this.stringBuilder.WriteString(fmt.Sprintf("0x%08x", v))
	} else {
		this.stringBuilder.WriteString(fmt.Sprintf("%v", v))
	}
}

func (this *descriptionContext) describeUint64(v uint64, asHex bool) {
	if asHex {
		this.stringBuilder.WriteString(fmt.Sprintf("0x%016x", v))
	} else {
		this.stringBuilder.WriteString(fmt.Sprintf("%v", v))
	}
}

func (this *descriptionContext) describeUint(v uint, asHex bool) {
	if asHex {
		this.stringBuilder.WriteString(fmt.Sprintf("0x%08x", v))
	} else {
		this.stringBuilder.WriteString(fmt.Sprintf("%v", v))
	}
}

func (this *descriptionContext) describeReflect(v reflect.Value, asHex bool) {
	// Descriptions are attempted in the following priority order:
	// - If it's a duplicate that has been seen before, print the reference name.
	// - If it's a duplicate that hasn't been seen yet, prepend the reference
	//   name, then describe normally.
	// - If there's a custom describer, use that.
	// - Describe normally.
	// The actual code logic ends up a little convoluted for performance reasons.

	// Special case: Follow reflect values if we can.
	if v.IsValid() && !v.IsZero() {
		if v.Type() == reflectValueType {
			this.stringBuilder.WriteString("reflect.Value(")
			if v.CanInterface() {
				this.describeReflect(v.Interface().(reflect.Value), asHex)
			} else if canDereferenceNestedReflectValues() {
				this.describeReflect(dereferenceNestedReflectValue(v), asHex)
			} else {
				this.stringBuilder.WriteString(fmt.Sprintf("%v", v))
			}
			this.stringBuilder.WriteString(")")
			return
		} else if v.Type() == reflectTypeType {
			this.stringBuilder.WriteString(fmt.Sprintf("reflect.Type(%v)", v))
			return
		}
	}

	switch v.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.Struct:
		var ptr uintptr
		if v.Kind() == reflect.Struct {
			if v.CanAddr() {
				ptr = v.Addr().Pointer()
			} else {
				break
			}
		} else {
			ptr = v.Pointer()
		}
		if referenceName, ok := this.referenceNames[ptr]; ok {
			if _, ok := this.seenReferences[ptr]; ok {
				this.stringBuilder.WriteString(referencePrefix)
				this.stringBuilder.WriteString(fmt.Sprintf("%v", referenceName))
				return
			}
			this.stringBuilder.WriteString(fmt.Sprintf("%v", referenceName))
			this.stringBuilder.WriteString(referenceSeparator)
			this.seenReferences[ptr] = true
		}
	}

	if v.IsValid() && !v.IsZero() {
		if customDescriber, ok := customDescribers[v.Type()]; ok && customDescriber != nil {
			this.stringBuilder.WriteString(customDescriber(v))
			return
		}
	}

	switch v.Kind() {
	case reflect.Interface:
		if v.IsNil() {
			this.stringBuilder.WriteString(nilPointer)
		} else {
			this.stringBuilder.WriteString(interfacePrefix)
			this.describeReflect(v.Elem(), false)
		}
	case reflect.Ptr:
		if v.IsNil() {
			this.stringBuilder.WriteString(nilPointer)
		} else {
			this.stringBuilder.WriteString(pointerPrefix)
			this.describeReflect(v.Elem(), false)
		}
	case reflect.Slice, reflect.Array:
		this.describeArray(v)
	case reflect.Map:
		this.describeMap(v)
	case reflect.Struct:
		this.describeStruct(v)
	case reflect.String:
		this.stringBuilder.WriteString(openString)
		this.stringBuilder.WriteString(v.String())
		this.stringBuilder.WriteString(closeString)
	case reflect.Func, reflect.Chan:
		// Do nothing
	case reflect.Uint8:
		this.describeUint8(uint8(v.Uint()), asHex)
	case reflect.Uint16:
		this.describeUint16(uint16(v.Uint()), asHex)
	case reflect.Uint32:
		this.describeUint32(uint32(v.Uint()), asHex)
	case reflect.Uint64:
		this.describeUint64(uint64(v.Uint()), asHex)
	case reflect.Uint:
		this.describeUint(uint(v.Uint()), asHex)
	default:
		this.stringBuilder.WriteString(fmt.Sprintf("%v", v))
	}
}

func (this *descriptionContext) describe(v interface{}) string {
	rv := reflect.ValueOf(v)
	this.referenceNames = findDuplicates(rv)
	this.stringBuilder.Reset()
	this.seenReferences = make(map[uintptr]bool)
	this.describeReflect(rv, false)
	return this.stringBuilder.String()
}

// ----------
// Public API
// ----------

// If enabled, panic on errors rather than returning an error string.
// This is only useful when debugging this library.
var PanicOnError bool = false

// If disabled, nested reflect.Value structures cannot be examined.
// This switch does nothing if compiled with `-tags safe`, or compiled for
// GopherJS or AppEngine, whereby unsafe operations won't even be compiled in.
var EnableUnsafeOperations = true

// Describes an object in a single line of text.
//
// * Basic types are printed as-is (as if printed via %v)
// * Strings are enclosed in quotes ""
// * Nil pointers are simply printed as nil
// * Pointers are preceded by &
// * Interfaces are preceded by @
// * Slices and arrays are enclosed in []
// * Slices and arrays of unsigned int types are printed as hex
// * Maps are enclosed in {}, listing keys and values separated by :
// * Structs are preceded by the struct type name, and fields are enclosed in (),
//   listing field names and values separated by :
// * Duplicate and cyclic references will be marked as follows:
//   - The first instance will be prepended by an ID and =
//   - Further instances will be replaced by a reference: $ and the ID
func Describe(v interface{}) (description string) {
	defer func() {
		if !PanicOnError {
			if e := recover(); e != nil {
				description = fmt.Sprintf("go-describe: Unexpected error: %v", e)
			}
		}
	}()

	context := descriptionContext{}
	description = context.describe(v)
	return
}

// User-injectable value describer. Pass a function like this to
// SetCustomDescriber() when you want different describe behavior from the default.
type CustomDescriber func(reflect.Value) string

// Add a custom describer for a data type.
//
// The describer function will be passed an instance of that type as a
// reflect.Value, and be expected to return a string that describes it.
//
// The convention is to print a recognizable type name, followed by the
// description enclosed in <>. For example: url<http://example.com>
//
// Note: Please pass in concrete types rather than pointer or interface types.
//
// Note: url.URL and time.Time already have custom describers, but you can
//       override them if you wish.
func SetCustomDescriber(t reflect.Type, describer CustomDescriber) {
	customDescribers[t] = describer
}
