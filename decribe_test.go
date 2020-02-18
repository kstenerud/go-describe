package describe

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

type InnerStruct struct {
	number int
}

type OuterStruct struct {
	AnInt          int
	PInt           *int
	Bytes          []byte
	URL            *url.URL
	Time           time.Time
	AStruct        InnerStruct
	PStruct        *InnerStruct
	AnotherPStruct *InnerStruct
	AMap           map[interface{}]interface{}
}

func TestDemonstration(t *testing.T) {
	urlVal, _ := url.Parse("http://example.com")
	intVal := 1
	structVal := InnerStruct{100}
	v := OuterStruct{
		AnInt:          4,
		PInt:           &intVal,
		Bytes:          []byte{0xff, 0x80, 0x44, 0x01},
		URL:            urlVal,
		Time:           time.Date(2020, time.Month(1), 1, 1, 1, 1, 0, time.UTC),
		AStruct:        InnerStruct{number: 200},
		PStruct:        &structVal,
		AnotherPStruct: nil,
		AMap: map[interface{}]interface{}{
			"flt":   1.5,
			"str":   "blah",
			"inner": InnerStruct{number: 99},
		},
	}
	fmt.Printf("%v\n", v)
	fmt.Printf("%v\n", Describe(v))

	// Can't compare the entire string because the map entries are in unpredictable order
	expectedPrefix := `OuterStruct(AnInt:4 PInt:&1 Bytes:[0xff 0x80 0x44 0x01] URL:&url<http://example.com> Time:time<2020-01-01 01:01:01 +0000 UTC> AStruct:InnerStruct(number:200) PStruct:&InnerStruct(number:100) AnotherPStruct:nil AMap:{`
	actual := Describe(v)
	if !strings.HasPrefix(actual, expectedPrefix) {
		t.Errorf("Expected %v to start with %v", actual, expectedPrefix)
	}
}

type RecursiveStruct struct {
	IntVal       int
	RecursivePtr *RecursiveStruct
	data         interface{}
}

func TestRecursive(t *testing.T) {
	someMap := make(map[string]interface{})
	someMap["mykey"] = someMap
	v1 := RecursiveStruct{}
	v1.IntVal = 100
	v1.RecursivePtr = &v1
	v1.data = someMap
	v2 := RecursiveStruct{}
	v2.IntVal = 5
	v2.RecursivePtr = &v1
	v2.data = someMap
	slice := []interface{}{&v1, &v2, &v1}

	fmt.Printf("%v\n", slice)
	fmt.Printf("%v\n", Describe(slice))

	expected1 := `[@&1=RecursiveStruct(IntVal:100 RecursivePtr:&$1 data:@2={"mykey":@$2}) @&RecursiveStruct(IntVal:5 RecursivePtr:&$1 data:@$2) @&$1]`
	expected2 := `[@&2=RecursiveStruct(IntVal:100 RecursivePtr:&$2 data:@1={"mykey":@$1}) @&RecursiveStruct(IntVal:5 RecursivePtr:&$2 data:@$1) @&$2]`
	actual := Describe(slice)
	if actual != expected1 && actual != expected2 {
		t.Errorf("Expected %v but got %v", expected1, actual)
	}
}

func TestRecursiveNotAddressable(t *testing.T) {
	v := RecursiveStruct{}
	v.RecursivePtr = &v

	expected := `RecursiveStruct(IntVal:0 RecursivePtr:&1=RecursiveStruct(IntVal:0 RecursivePtr:&$1 data:nil) data:nil)`
	actual := Describe(v)
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestRecursiveMap(t *testing.T) {
	v := make(map[string]interface{})
	v["mykey"] = v
	expected := `1={"mykey":@$1}`
	actual := Describe(v)
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestRecursiveMapInStruct(t *testing.T) {
	someMap := make(map[string]interface{})
	someMap["mykey"] = someMap
	v := RecursiveStruct{}
	v.data = someMap

	// Uncommenting this would result in a stack overflow
	// fmt.Printf("%v\n", v)
	fmt.Printf("%v\n", Describe(v))

	expected := `RecursiveStruct(IntVal:0 RecursivePtr:nil data:@1={"mykey":@$1})`
	actual := Describe(v)
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestReflectValue(t *testing.T) {
	v := 1
	rv := reflect.ValueOf(v)

	expected := `reflect.Value(1)`
	actual := Describe(rv)
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}
