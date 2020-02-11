package describe

import (
	"fmt"
	"reflect"
	"strings"
)

func describeArray(v reflect.Value) string {
	var builder strings.Builder
	asHex := false
	switch v.Type().Elem().Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		asHex = true
	}
	isFirst := true
	builder.WriteString("[")
	for i := 0; i < v.Len(); i++ {
		if isFirst {
			isFirst = false
		} else {
			builder.WriteString(" ")
		}
		builder.WriteString(describeRV(v.Index(i), asHex))
	}
	builder.WriteString("]")
	return builder.String()
}

func describeMap(v reflect.Value) string {
	var builder strings.Builder
	isFirst := true
	builder.WriteString("{")
	iter := v.MapRange()
	for iter.Next() {
		if isFirst {
			isFirst = false
		} else {
			builder.WriteString(" ")
		}
		builder.WriteString(describeRV(iter.Key(), false))
		builder.WriteString(":")
		builder.WriteString(describeRV(iter.Value(), false))
	}
	builder.WriteString("}")
	return builder.String()
}

func describeStruct(v reflect.Value) string {
	var builder strings.Builder
	isFirst := true
	builder.WriteString(v.Type().Name())
	builder.WriteString("(")
	for i := 0; i < v.NumField(); i++ {
		if isFirst {
			isFirst = false
		} else {
			builder.WriteString(" ")
		}
		builder.WriteString(v.Type().Field(i).Name)
		builder.WriteString(":")
		builder.WriteString(describeRV(v.Field(i), false))
	}
	builder.WriteString(")")
	return builder.String()
}

func describeUint8(v uint8, asHex bool) string {
	if asHex {
		return fmt.Sprintf("0x%02x", v)
	}
	return fmt.Sprintf("%v", v)
}

func describeUint16(v uint16, asHex bool) string {
	if asHex {
		return fmt.Sprintf("0x%04x", v)
	}
	return fmt.Sprintf("%v", v)
}

func describeUint32(v uint32, asHex bool) string {
	if asHex {
		return fmt.Sprintf("0x%08x", v)
	}
	return fmt.Sprintf("%v", v)
}

func describeUint64(v uint64, asHex bool) string {
	if asHex {
		return fmt.Sprintf("0x%016x", v)
	}
	return fmt.Sprintf("%v", v)
}

func describeUint(v uint, asHex bool) string {
	if asHex {
		return fmt.Sprintf("0x%08x", v)
	}
	return fmt.Sprintf("%v", v)
}

func describeRV(v reflect.Value, asHex bool) string {
	switch v.Kind() {
	case reflect.Interface:
		if v.IsNil() {
			return "nil"
		}
		return "@" + describeRV(v.Elem(), false)
	case reflect.Ptr:
		if v.IsNil() {
			return "nil"
		}
		return "&" + describeRV(v.Elem(), false)
	case reflect.Slice, reflect.Array:
		return describeArray(v)
	case reflect.Map:
		return describeMap(v)
	case reflect.Struct:
		return describeStruct(v)
	case reflect.String:
		return "\"" + v.String() + "\""
	case reflect.Func, reflect.Chan:
		return ""
	case reflect.Uint8:
		return describeUint8(uint8(v.Uint()), asHex)
	case reflect.Uint16:
		return describeUint16(uint16(v.Uint()), asHex)
	case reflect.Uint32:
		return describeUint32(uint32(v.Uint()), asHex)
	case reflect.Uint64:
		return describeUint64(uint64(v.Uint()), asHex)
	case reflect.Uint:
		return describeUint(uint(v.Uint()), asHex)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// Describes an object in a single line of text.
// - Basic types are printed as-is
// - Strings are enclosed in quotes ""
// - Pointers are preceded by &
// - Interfaces are preceded by @
// - Slices and arrays are enclosed in []
// - Slices with unsigned int types are printed as hex
// - Maps are enclosed in {}
// - Structs are preceded by the struct type name, and fields are enclosed in ()
func Describe(v interface{}) string {
	return describeRV(reflect.ValueOf(v), false)
}
