// Package describe provides an API for describing objects as a single line or
// multiline. It provides MUCH more information than the `%v` formatter does,
// allowing you to see more about complex objects at a glance.
//
// It handles recursive data, and can describe structures that would cause `%v`
// to stack overflow.
//
// The description is structured as follows:
//
// * Basic types are printed the same as by `%v`
// * Strings are enclosed in quotes `""`
// * Non-nil pointers are prefixed with `*`
// * Nil pointers are printed as `nil`
// * Interfaces are prefixed with `@`
// * The empty interface type is printed as `interface` (not `interface{}`)
// * Slices and arrays are preceded by a type, and enclosed in `[]`
// * Slices and arrays of unsigned int types are printed as hex
// * Maps begin with `key_type:value_type`, with elements enclosed in `{}`.
//   Key-value pairs separated by `=`
// * Structs are preceded by a type, with elements enclosed in `<>`.
//   Field-value pairs are separated by `=`
// * Functions begin with `func`, with in and out params enclosed in `()`.
//   Example: `func(int, bool)(string, bool)`
// * Nil functions begin with `nilfunc`. Example: `nilfunc(int)(string)`
// * Unidirectional channels are printed as `<-chan type` and `chan<- type`
// * Bidirectional channels are printed as `chan<sometype>`
// * Uintptr and UnsafePointer are printed as hex, in the width of the host system
// * Invalid values are printed as `invalid`
// * Custom describers by convention print a type name, then a description within
//   `<>`. Example: `url.URL<http://xyz.com>`
// * Duplicate and cyclic data will be marked as follows:
//   - The first instance is prefixed by a unique numeric reference ID, then `~`
//   - Further instances are replaced by `$`, then the referenced ID
//
// Note: Only data is printed; type-specific things such as methods are not.
//
// Note: describe uses the `unsafe` package to expose unexported
//       `reflect.Value` and `reflect.Type` objects. This functionality can
//       be disabled by compiling with `-tags safe`, or by setting
//       `describe.EnableUnsafeOperations` to `false`. It will be
//       automatically disabled if compiling for GopherJS or AppEngine.
package describe

import (
	"bytes"
	"fmt"
	"net/url"
	"reflect"
	"sync"
	"time"
)

// ------------
// Package Init
// ------------

func init() {
	initUnsafe()

	SetCustomDescriber(reflect.TypeOf(url.URL{}), describeURL)
	SetCustomDescriber(reflect.TypeOf(time.Time{}), describeStringer)
}

// ----------------
// Global Constants
// ----------------

const maxIndentStep = 100

const (
	tokOpenString             = `"`
	tokCloseString            = `"`
	tokOpenArray              = "["
	tokCloseArray             = "]"
	tokOpenMap                = "{"
	tokCloseMap               = "}"
	tokOpenStruct             = "<"
	tokCloseStruct            = ">"
	tokOpenFunc               = "("
	tokCloseFunc              = ")"
	tokItemSeparator          = " "
	tokItemSeparatorMultiline = "\n"
	tokIndent                 = " "
	tokMapTypeSeparator       = ":"
	tokKeyValueSeparator      = "="
	tokPointerPrefix          = "*"
	tokInterfacePrefix        = "@"
	tokReferenceSeparator     = "~"
	tokReferencePrefix        = "$"
	tokNilPointer             = "nil"
	tokEmptyInterface         = "interface"
	tokInvalid                = "invalid"
	tokRecvChannel            = "<-chan"
	tokSendChannel            = "chan<-"
)

const is64BitUint = uint64(^uint(0)) == ^uint64(0)
const is64BitUintptr = uint64(^uintptr(0)) == ^uint64(0)

var reflectValueType = reflect.ValueOf(reflect.ValueOf(true)).Type()
var reflectTypeType = reflect.TypeOf((*reflect.Type)(nil)).Elem()
var emptyInterfaceType = reflect.ValueOf([]interface{}{}).Type().Elem()

// -----------
// Global Data
// -----------

var customDescribers sync.Map

// ---------
// Utilities
// ---------

// If this gets called, there's a bug in this library.
// Setting `DebugPanics = true` might help track down the cause.
//
// Please send bug reports to https://github.com/kstenerud/go-describe/issues
func notifyLibraryBug(format string, params ...interface{}) string {
	description := fmt.Sprintf(format, params...)
	return fmt.Sprintf("go-describe.BUG(%v)", description)
}

func isNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		return v.IsNil()
	}
	return false
}

func getTypeName(t reflect.Type) string {
	if t == emptyInterfaceType {
		return tokEmptyInterface
	}

	if t.Kind() == reflect.Chan {
		nameBytes := []byte(fmt.Sprintf("%v", t))
		index := bytes.IndexByte(nameBytes, byte(' '))
		if index < 0 {
			return notifyLibraryBug("could not parse chan type %v", string(nameBytes))
		}
		typeName := string(nameBytes[index+1:])

		if t.ChanDir()&reflect.BothDir == reflect.BothDir {
			return fmt.Sprintf("chan%v%v%v", tokOpenStruct, typeName, tokCloseStruct)
		}

		chanDir := tokRecvChannel
		if t.ChanDir()&reflect.SendDir != 0 {
			chanDir = tokSendChannel
		}
		return fmt.Sprintf("%v %v", chanDir, typeName)
	}

	return fmt.Sprintf("%v", t)
}

func stringifyUint(value uint64) string {
	if is64BitUint {
		return fmt.Sprintf("0x%016x", value)
	}
	return fmt.Sprintf("0x%08x", value)
}

func stringifyAddress(address uint64) string {
	if is64BitUintptr {
		return fmt.Sprintf("0x%016x", address)
	}
	return fmt.Sprintf("0x%08x", address)
}

func getInterfaceAsReflectValue(v reflect.Value) (value reflect.Value, ok bool) {
	if v.CanInterface() {
		return v.Interface().(reflect.Value), true
	}
	if canExposeInterface() {
		return exposeInterface(v).(reflect.Value), true
	}
	return v, false
}

func getInterfaceAsReflectType(v reflect.Value) (t reflect.Type, ok bool) {
	if v.CanInterface() {
		return v.Interface().(reflect.Type), true
	}
	if canExposeInterface() {
		return exposeInterface(v).(reflect.Type), true
	}
	return v.Type(), false
}

// -----------------
// Custom Describers
// -----------------

func runCustomDescriber(v reflect.Value, describer CustomDescriber) (description string) {
	// A custom describer runs unknown user-supplied code that we don't control.
	// If it panics, return the stringified contents of the panic instead.
	defer func() {
		// Allow panic to escape if debugging
		if !DebugPanics {
			if e := recover(); e != nil {
				description = fmt.Sprintf("panic(%v)", e)
			}
		}
	}()

	description = describer(v)
	return
}

func describeURL(v reflect.Value) string {
	asString := v.Interface().(url.URL)
	return fmt.Sprintf(`%v%v%v%v`, v.Type(), tokOpenStruct, asString.String(), tokCloseStruct)
}

func describeStringer(v reflect.Value) string {
	asString := v.Interface().(fmt.Stringer)
	return fmt.Sprintf(`%v%v%v%v`, v.Type(), tokOpenStruct, asString.String(), tokCloseStruct)
}

// -----------------
// Duplicates Finder
// -----------------

func checkForDuplicate(seenPointerCounts map[uintptr]int, ptr uintptr) (duplicateFound bool) {
	if count, ok := seenPointerCounts[ptr]; ok {
		seenPointerCounts[ptr] = count + 1
		duplicateFound = true
		return
	}

	seenPointerCounts[ptr] = 1
	return
}

func findDuplicatesRecursive(seenPointerCounts map[uintptr]int, v reflect.Value) {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		findDuplicatesRecursive(seenPointerCounts, v.Elem())
	case reflect.Array:
		if v.CanAddr() {
			if checkForDuplicate(seenPointerCounts, v.Addr().Pointer()) {
				return
			}
		}
		for i := 0; i < v.Len(); i++ {
			findDuplicatesRecursive(seenPointerCounts, v.Index(i))
		}
	case reflect.Slice:
		if checkForDuplicate(seenPointerCounts, v.Pointer()) {
			return
		}
		for i := 0; i < v.Len(); i++ {
			findDuplicatesRecursive(seenPointerCounts, v.Index(i))
		}
	case reflect.Map:
		if checkForDuplicate(seenPointerCounts, v.Pointer()) {
			return
		}
		for iter := mapRange(v); iter.Next(); {
			findDuplicatesRecursive(seenPointerCounts, iter.Value())
		}
	case reflect.Struct:
		if v.CanAddr() && checkForDuplicate(seenPointerCounts, v.Addr().Pointer()) {
			return
		}
		for i := 0; i < v.NumField(); i++ {
			findDuplicatesRecursive(seenPointerCounts, v.Field(i))
		}
	}
}

func findDuplicates(v reflect.Value) map[uintptr]int {
	seenPointerCounts := map[uintptr]int{}
	findDuplicatesRecursive(seenPointerCounts, v)
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

// ---------
// Describer
// ---------

func (this *describer) increaseIndent() {
	this.currentIndent += this.indentStep
}

func (this *describer) decreaseIndent() {
	this.currentIndent -= this.indentStep
}

func (this *describer) writeString(value string) {
	this.stringBuilder.WriteString(value)
}

func (this *describer) writeFmt(format string, args ...interface{}) {
	this.writeString(fmt.Sprintf(format, args...))
}

func (this *describer) writeItemSeparator(isFirst bool) {
	if this.indentStep > 0 {
		this.writeString(tokItemSeparatorMultiline)
		for i := 0; i < this.currentIndent; i++ {
			this.writeString(tokIndent)
		}
	} else if !isFirst {
		this.writeString(tokItemSeparator)
	}
}

func (this *describer) writeKeyValueSeparator() {
	if this.indentStep > 0 {
		this.writeFmt(" %v ", tokKeyValueSeparator)
	} else {
		this.writeString(tokKeyValueSeparator)
	}
}

func (this *describer) describeArray(v reflect.Value) {
	isInUnsignedArray := false
	switch v.Type().Elem().Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		isInUnsignedArray = true
	}
	this.writeString(getTypeName(v.Type().Elem()))
	this.writeString(tokOpenArray)
	this.increaseIndent()
	isFirst := true
	for i := 0; i < v.Len(); i++ {
		this.writeItemSeparator(isFirst)
		isFirst = false
		this.describeReflectedValue(v.Index(i), isInUnsignedArray)
	}
	this.decreaseIndent()
	this.writeItemSeparator(true)
	this.writeString(tokCloseArray)
}

func (this *describer) describeMap(v reflect.Value) {
	this.writeString(getTypeName(v.Type().Key()))
	this.writeString(tokMapTypeSeparator)
	this.writeString(getTypeName(v.Type().Elem()))
	this.writeString(tokOpenMap)
	this.increaseIndent()
	isFirst := true
	for iter := mapRange(v); iter.Next(); {
		this.writeItemSeparator(isFirst)
		isFirst = false
		this.describeReflectedValue(iter.Key(), false)
		this.writeKeyValueSeparator()
		this.describeReflectedValue(iter.Value(), false)
	}
	this.decreaseIndent()
	this.writeItemSeparator(true)
	this.writeString(tokCloseMap)
}

func (this *describer) describeStruct(v reflect.Value) {
	this.writeString(getTypeName(v.Type()))
	this.writeString(tokOpenStruct)
	this.increaseIndent()
	isFirst := true
	for i := 0; i < v.NumField(); i++ {
		this.writeItemSeparator(isFirst)
		isFirst = false
		this.writeString(v.Type().Field(i).Name)
		this.writeKeyValueSeparator()
		this.describeReflectedValue(v.Field(i), false)
	}
	this.decreaseIndent()
	this.writeItemSeparator(true)
	this.writeString(tokCloseStruct)
}

func (this *describer) describeFunc(v reflect.Value) {
	var t reflect.Type = v.Type()

	name := "func"
	if v.IsNil() {
		name = "nilfunc"
	}
	this.writeString(name)
	this.writeString(tokOpenFunc)
	numIn := t.NumIn()
	for i := 0; i < numIn; i++ {
		this.writeString(getTypeName(t.In(i)))
		if i < numIn-1 {
			this.writeString(", ")
		}
	}
	this.writeString(tokCloseFunc)
	this.writeString(tokOpenFunc)
	numOut := t.NumOut()
	for i := 0; i < numOut; i++ {
		this.writeString(getTypeName(t.Out(i)))
		if i < numOut-1 {
			this.writeString(", ")
		}
	}
	this.writeString(tokCloseFunc)
}

func (this *describer) describeUint8(v uint8, isInUnsignedArray bool) {
	if isInUnsignedArray {
		this.writeFmt("0x%02x", v)
	} else {
		this.writeFmt("%v", v)
	}
}

func (this *describer) describeUint16(v uint16, isInUnsignedArray bool) {
	if isInUnsignedArray {
		this.writeFmt("0x%04x", v)
	} else {
		this.writeFmt("%v", v)
	}
}

func (this *describer) describeUint32(v uint32, isInUnsignedArray bool) {
	if isInUnsignedArray {
		this.writeFmt("0x%08x", v)
	} else {
		this.writeFmt("%v", v)
	}
}

func (this *describer) describeUint64(v uint64, isInUnsignedArray bool) {
	if isInUnsignedArray {
		this.writeFmt("0x%016x", v)
	} else {
		this.writeFmt("%v", v)
	}
}

func (this *describer) describeUint(v uint, isInUnsignedArray bool) {
	if isInUnsignedArray {
		this.writeString(stringifyUint(uint64(v)))
	} else {
		this.writeFmt("%v", v)
	}
}

func (this *describer) tryDescribeNil(v reflect.Value) (didDescribeNil bool) {
	if isNil(v) {
		if v.Kind() == reflect.Func {
			this.describeFunc(v)
		} else {
			this.writeString(tokNilPointer)
		}
		didDescribeNil = true
		return
	}

	didDescribeNil = false
	return
}

func (this *describer) tryDescribeReflectValueOrType(v reflect.Value) (didDescribe bool) {
	// reflect.Value and reflect.Type field contents aren't always accessible
	// depending on their visibility, so we must handle them specially.

	if !v.IsValid() {
		didDescribe = false
		return
	}

	if v.Type() == reflectValueType {
		this.writeString("reflect.Value")
		this.writeString(tokOpenStruct)
		if rValue, ok := getInterfaceAsReflectValue(v); ok {
			this.describeReflectedValue(rValue, false)
		} else {
			this.writeFmt("%v", v)
		}
		this.writeString(tokCloseStruct)
		didDescribe = true
		return
	}

	if v.Type().Implements(reflectTypeType) {
		this.writeString("reflect.Type")
		this.writeString(tokOpenStruct)
		if rValue, ok := getInterfaceAsReflectType(v); ok {
			this.writeString(getTypeName(rValue))
		} else {
			this.writeFmt("%v", v)
		}
		this.writeString(tokCloseStruct)
		didDescribe = true
		return
	}

	didDescribe = false
	return
}

func (this *describer) tryDescribeReference(v reflect.Value) (didReplaceWithReference bool) {
	// Note: This method has the side effect of modifying this.seenReferences

	switch v.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.Struct:
		var ptr uintptr
		if v.Kind() == reflect.Struct || v.Kind() == reflect.Array {
			if !v.CanAddr() {
				didReplaceWithReference = false
				return
			}
			ptr = v.Addr().Pointer()
		} else {
			ptr = v.Pointer()
		}
		if referenceName, ok := this.referenceNames[ptr]; ok {
			if _, ok := this.seenReferences[ptr]; ok {
				// The first instance of a repeated structure was described
				// already, so we replace with a reference.
				this.writeString(tokReferencePrefix)
				this.writeFmt("%v", referenceName)
				didReplaceWithReference = true
				return
			}

			// We're only marking the first instance of a repeated structure
			// rather than replacing it, so in this case we haven't replaced
			// with a reference.
			this.writeFmt("%v", referenceName)
			this.writeString(tokReferenceSeparator)
			this.seenReferences[ptr] = true
			didReplaceWithReference = false
			return
		}
	}

	didReplaceWithReference = false
	return
}

func (this *describer) tryUseCustomDescriber(v reflect.Value) (didUseCustomDescriber bool) {
	if !v.IsValid() {
		didUseCustomDescriber = false
		return
	}

	if customDescriber, ok := customDescribers.Load(v.Type()); ok && customDescriber != nil {
		this.writeString(runCustomDescriber(v, customDescriber.(CustomDescriber)))
		didUseCustomDescriber = true
		return
	}

	didUseCustomDescriber = false
	return
}

func (this *describer) describeNormally(v reflect.Value, isInUnsignedArray bool) {
	switch v.Kind() {
	case reflect.Bool, reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Complex64, reflect.Complex128:
		this.writeFmt("%v", v)
	case reflect.Uint8:
		this.describeUint8(uint8(v.Uint()), isInUnsignedArray)
	case reflect.Uint16:
		this.describeUint16(uint16(v.Uint()), isInUnsignedArray)
	case reflect.Uint32:
		this.describeUint32(uint32(v.Uint()), isInUnsignedArray)
	case reflect.Uint64:
		this.describeUint64(uint64(v.Uint()), isInUnsignedArray)
	case reflect.Uint:
		this.describeUint(uint(v.Uint()), isInUnsignedArray)
	case reflect.String:
		this.writeString(tokOpenString)
		this.writeString(v.String())
		this.writeString(tokCloseString)
	case reflect.Slice, reflect.Array:
		this.describeArray(v)
	case reflect.Map:
		this.describeMap(v)
	case reflect.Struct:
		this.describeStruct(v)
	case reflect.Interface:
		this.writeString(tokInterfacePrefix)
		this.describeReflectedValue(v.Elem(), false)
	case reflect.Ptr:
		this.writeString(tokPointerPrefix)
		this.describeReflectedValue(v.Elem(), false)
	case reflect.Uintptr:
		this.writeString(stringifyAddress(v.Uint()))
	case reflect.UnsafePointer:
		this.writeString(tokPointerPrefix)
		this.writeString(stringifyAddress(uint64(v.UnsafeAddr())))
	case reflect.Invalid:
		this.writeString(tokInvalid)
	case reflect.Func:
		this.describeFunc(v)
	case reflect.Chan:
		this.writeString(getTypeName(v.Type()))
	default:
		this.writeString(notifyLibraryBug("unhandled type %v (kind %v): %v", v.Type(), v.Kind(), v))
	}
}

func (this *describer) describeReflectedValue(v reflect.Value, isInsideUnsignedArray bool) {
	if this.tryDescribeNil(v) {
		return
	}

	if this.tryDescribeReflectValueOrType(v) {
		return
	}

	if this.tryDescribeReference(v) {
		return
	}

	if this.tryUseCustomDescriber(v) {
		return
	}

	this.describeNormally(v, isInsideUnsignedArray)
}

func (this *describer) sanityCheck() {
	if this.indentStep > maxIndentStep {
		panic(fmt.Errorf("Sanity check fail: indent step %v > max of %v", this.indentStep, maxIndentStep))
	}
}

func (this *describer) describe(v interface{}) (description string) {
	defer func() {
		// Allow panic to escape if debugging
		if !DebugPanics {
			if e := recover(); e != nil {
				description = notifyLibraryBug("%v", e)
			}
		}
	}()

	this.sanityCheck()

	rv := reflect.ValueOf(v)
	this.referenceNames = findDuplicates(rv)
	this.stringBuilder.Reset()
	this.seenReferences = make(map[uintptr]bool)
	this.describeReflectedValue(rv, false)
	description = this.stringBuilder.String()
	return
}

// ----------
// Public API
// ----------

// If enabled, allow panics to bubble up instead of returning an error string.
// This is useful for tracing the cause of the panic.
var DebugPanics bool = false

// If disabled, nested reflect.Value structures cannot be examined.
// This switch does nothing if compiled with `-tags safe` or if compiled for
// GopherJS or AppEngine, whereby unsafe operations won't even be compiled in.
var EnableUnsafeOperations = true

// Describes an object in a single line or multiple lines of text. See package
// description for information about how data is represented.
//
// if indentStep > 0, it will print in multiline mode, indenting that number of
// spaces when it enters a struct/map/array/slice.
func Describe(v interface{}, indentStep int) (description string) {
	context := describer{}
	context.indentStep = indentStep
	description = context.describe(v)
	return
}

// Alias to `Describe(v, 0)`. Call `describe.D(myobject)` to get a one-line
// description for logging, debugging, etc.
func D(v interface{}) (description string) {
	return Describe(v, 0)
}

// User-defined value describer. Pass to SetCustomDescriber() when you want
// different behavior from the default.
type CustomDescriber func(reflect.Value) string

// Add a custom describer for a data type.
//
// describer will be passed an instance of type t as a reflect.Value, and be
// expected to return a string that describes it.
//
// If a custom describer panics, the stringified contents of the panic will be
// used instead (unless DebugPanics is true).
//
// The convention is to print the type name, followed by a description
// enclosed in <>. Example: net.URL<http://example.com>
//
// Passing a nil describer will disable the custom describer for that type.
//
// Note: t should be a concrete type rather than a pointer or interface type.
//
// Note: url.URL and time.Time already have custom describers by default, but
//       you can override or disable them if you wish.
func SetCustomDescriber(t reflect.Type, describer CustomDescriber) {
	customDescribers.Store(t, describer)
}
